package core

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/rpc"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/toolkits/pkg/logger"
	"github.com/ugorji/go/codec"

	"github.com/didi/nightingale/src/dataobj"
	"github.com/didi/nightingale/src/model"
	"github.com/didi/nightingale/src/modules/collector/cache"
	"github.com/didi/nightingale/src/toolkits/address"
	"github.com/didi/nightingale/src/toolkits/identity"
)

// openfalcon v1/push 接口：指标转换，补充必要的标签数据
func PushV1(metricItems []*dataobj.MetricValue) error {
	var err error
	var items []*dataobj.MetricValue
	now := time.Now().Unix()
	filterStr := cache.GarbageCache.Get()

	for _, item := range metricItems {
		logger.Debug("->recv: ", item)
		if item.Endpoint == "" {
			item.Endpoint = identity.Identity
		}
		err = item.ConvertAndCheckValidity(filterStr, now)
		if err != nil {
			logger.Debugf("metric:%v err:%v", item, err)
			// 如果数据有问题，直接跳过吧，比如mymon采集的到的数据，其实只有一个有问题，剩下的都没问题
			continue
		}
		// 指标转换：白名单、渲染标签
		item, err := convertMetricItem(item)
		if err != nil {
			logger.Debugf("metric:%v err:%v", item, err)
			continue
		}
		if item.CounterType == dataobj.COUNTER {
			item = CounterToGauge(item)
			if item == nil {
				continue
			}
		}
		if item.CounterType == dataobj.SUBTRACT {
			item = SubtractToGauge(item)
			if item == nil {
				continue
			}
		}
		logger.Debug("push item: ", item)
		items = append(items, item)
	}

	addrs := address.GetRPCAddresses("transfer")
	count := len(addrs)
	retry := 0
	for {
		for _, i := range rand.Perm(count) {
			addr := addrs[i]
			reply, err := rpcCall(addr, items)
			if err != nil {
				logger.Error(err)
				continue
			} else {
				if reply.Msg != "ok" {
					err = fmt.Errorf("some item push err: %s", reply.Msg)
					logger.Error(err)
				}
				return err
			}
		}

		time.Sleep(time.Millisecond * 500)

		retry += 1
		if retry == 3 {
			break
		}
	}

	return err
}

// 指标转换，补充必要的标签数据
func Push(metricItems []*dataobj.MetricValue) error {
	var err error
	var items []*dataobj.MetricValue
	now := time.Now().Unix()
	filterStr := cache.GarbageCache.Get()

	for _, item := range metricItems {
		logger.Debug("->recv: ", item)
		if item.Endpoint == "" {
			item.Endpoint = identity.Identity
		}
		err = item.ConvertAndCheckValidity(filterStr, now)
		if err != nil {
			logger.Debugf("metric:%v err:%v", item, err)
			// 如果数据有问题，直接跳过吧，比如mymon采集的到的数据，其实只有一个有问题，剩下的都没问题
			continue
		}
		if item.CounterType == dataobj.COUNTER {
			item = CounterToGauge(item)
			if item == nil {
				continue
			}
		}
		if item.CounterType == dataobj.SUBTRACT {
			item = SubtractToGauge(item)
			if item == nil {
				continue
			}
		}
		logger.Debug("push item: ", item)
		items = append(items, item)
	}

	addrs := address.GetRPCAddresses("transfer")
	count := len(addrs)
	retry := 0
	for {
		for _, i := range rand.Perm(count) {
			addr := addrs[i]
			reply, err := rpcCall(addr, items)
			if err != nil {
				logger.Error(err)
				continue
			} else {
				if reply.Msg != "ok" {
					err = fmt.Errorf("some item push err: %s", reply.Msg)
					logger.Error(err)
				}
				return err
			}
		}

		time.Sleep(time.Millisecond * 500)

		retry += 1
		if retry == 3 {
			break
		}
	}

	return err
}

func convertMetricItem(item *dataobj.MetricValue) (*dataobj.MetricValue, error) {
	//指标白名单
	metricInfo, exists := cache.MetricInfoCache.Get(item.Metric)
	if !exists {
		return nil, fmt.Errorf("metric:%v not exists in metricInfo", item)
	}

	switch metricInfo.EndpointType {
	case model.EndpointTypeInstance:
		index := strings.LastIndex(item.Endpoint, "_inst.")
		if index < 0 {
			return nil, fmt.Errorf("metric %s is not exists in metricInfo ", item.Metric)
		}
		uuid := item.Endpoint[index+6:]
		instance, exists := cache.AppInstanceCache.Get(uuid)
		if !exists {
			return nil, fmt.Errorf("instance %s is not found in instance Cache ", item.Endpoint)
		}
		item.TagsMap["app"] = instance.App
		item.TagsMap["group"] = instance.Group
		item.TagsMap["env"] = instance.Env
		// 如果指标本身不上报port,并且cmdb中存在端口信息，添加此标签
		if instance.Port > 0 {
			item.TagsMap["port"] = strconv.Itoa(instance.Port)
		}
		item.Endpoint = instance.Ident
	case model.EndpointTypePm:
		fallthrough
	case model.EndpointTypeDocker:
		fallthrough
	case model.EndpointTypeNetwork:
		ident := item.Endpoint
		endpointItem, exists := cache.EndpointCache.Get(ident)
		if !exists {
			return nil, fmt.Errorf("ident %s is not exists in endpoint cache", ident)
		}
		item.Endpoint = endpointItem.Ident
	default:
		// 其他类型丢弃
		return nil, fmt.Errorf("metric type is not found.item :%v", metricInfo)
	}
	item.Tags = dataobj.SortedTags(item.TagsMap)
	return item, nil
}

func rpcCall(addr string, items []*dataobj.MetricValue) (dataobj.TransferResp, error) {
	var reply dataobj.TransferResp
	var err error

	client := rpcClients.Get(addr)
	if client == nil {
		client, err = rpcClient(addr)
		if err != nil {
			return reply, err
		}
		affected := rpcClients.Put(addr, client)
		if !affected {
			defer func() {
				// 我尝试把自己这个client塞进map失败，说明已经有一个client塞进去了，那我自己用完了就关闭
				client.Close()
			}()

		}
	}

	timeout := time.Duration(8) * time.Second
	done := make(chan error, 1)

	go func() {
		err := client.Call("Transfer.Push", items, &reply)
		done <- err
	}()

	select {
	case <-time.After(timeout):
		logger.Warningf("rpc call timeout, transfer addr: %s\n", addr)
		rpcClients.Put(addr, nil)
		client.Close()
		return reply, fmt.Errorf("%s rpc call timeout", addr)
	case err := <-done:
		if err != nil {
			rpcClients.Del(addr)
			client.Close()
			return reply, fmt.Errorf("%s rpc call done, but fail: %v", addr, err)
		}
	}

	return reply, nil
}

func rpcClient(addr string) (*rpc.Client, error) {
	conn, err := net.DialTimeout("tcp", addr, time.Second*3)
	if err != nil {
		err = fmt.Errorf("dial transfer %s fail: %v", addr, err)
		logger.Error(err)
		return nil, err
	}

	var bufConn = struct {
		io.Closer
		*bufio.Reader
		*bufio.Writer
	}{conn, bufio.NewReader(conn), bufio.NewWriter(conn)}

	var mh codec.MsgpackHandle
	mh.MapType = reflect.TypeOf(map[string]interface{}(nil))

	rpcCodec := codec.MsgpackSpecRpc.ClientCodec(bufConn, &mh)
	client := rpc.NewClientWithCodec(rpcCodec)
	return client, nil
}

func CounterToGauge(item *dataobj.MetricValue) *dataobj.MetricValue {
	key := item.PK()

	old, exists := cache.MetricHistory.Get(key)
	cache.MetricHistory.Set(key, *item)

	if !exists {
		logger.Debugf("not found old item:%v, maybe this is the first item", item)
		return nil
	}

	if old.Value > item.Value {
		logger.Warningf("item:%v old value:%v greater than new value:%v", item, old.Value, item.Value)
		return nil
	}

	if old.Timestamp >= item.Timestamp {
		logger.Warningf("item:%v old timestamp:%v greater than new timestamp:%v", item, old.Timestamp, item.Timestamp)
		return nil
	}

	item.ValueUntyped = (item.Value - old.Value) / float64(item.Timestamp-old.Timestamp)
	item.CounterType = dataobj.GAUGE
	return item
}

func SubtractToGauge(item *dataobj.MetricValue) *dataobj.MetricValue {
	key := item.PK()

	old, exists := cache.MetricHistory.Get(key)
	cache.MetricHistory.Set(key, *item)

	if !exists {
		logger.Debugf("not found old item:%v, maybe this is the first item", item)
		return nil
	}

	if old.Timestamp >= item.Timestamp {
		logger.Warningf("item:%v old timestamp:%v greater than new timestamp:%v", item, old.Timestamp, item.Timestamp)
		return nil
	}

	if old.Timestamp <= item.Timestamp-2*item.Step {
		logger.Warningf("item:%v old timestamp:%v too old <= %v = (new timestamp: %v - 2 * step: %v), maybe some point lost", item, old.Timestamp, item.Timestamp-2*item.Step, item.Timestamp, item.Step)
		return nil
	}

	item.ValueUntyped = item.Value - old.Value
	item.CounterType = dataobj.GAUGE
	return item
}
