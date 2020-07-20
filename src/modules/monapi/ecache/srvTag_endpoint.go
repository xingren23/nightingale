package ecache

import (
	"encoding/json"
	"fmt"
	"github.com/didi/nightingale/src/modules/monapi/dataobj"
	"github.com/didi/nightingale/src/modules/monapi/redisc"
	"github.com/toolkits/pkg/logger"
	"strconv"
)

// 服务树节点串 + 资源类型 -> endpoint列表
func GetEndpointKeysFromRedis() ([]string, error) {
	ret := []string{}
	var cursor int
	batch := 20
	for {
		key := fmt.Sprintf("%s%s*", dataobj.EndpointKeyPrefix, dataobj.EndpointKeyDot)
		data, err := redisc.SCAN(cursor, key, batch)
		if err != nil {
			return ret, nil
		}
		cursor, err = strconv.Atoi(string(data[0].([]uint8)))
		if err != nil {
			return ret, nil
		}
		for _, d := range data[1].([]interface{}) {
			key := string(d.([]uint8))
			ret = append(ret, key)
		}
		if cursor == 0 {
			break
		}
	}

	return ret, nil
}

func GetEndpointByKeyFromRedis(srvType, srvTag string) ([]*dataobj.TagEndpoint, error) {
	ret := []*dataobj.TagEndpoint{}
	key := dataobj.BuildKey(srvType, srvTag)
	bs, err := redisc.SMEMBERS(key)
	if err != nil {
		return ret, err
	}
	if len(bs) == 0 {
		logger.Debugf("key [%s] not in redis cache.", key)
		// 不存在直接调用服务树接口
		switch srvType {
		case dataobj.EndpointKeyDocker:
			fallthrough
		case dataobj.EndpointKeyPM:
			res, err := dataobj.GetTreeByPage(srvTag, dataobj.CmdbSourceHost)
			if err != nil {
				return ret, err
			}
			for _, h := range res.Hosts {
				tagEndpoint := &dataobj.TagEndpoint{
					Ip:       h.Ip,
					HostName: h.HostName,
					EnvCode:  h.EnvCode,
				}
				ret = append(ret, tagEndpoint)
			}
		case dataobj.EndpointKeyNetwork:
			res, err := dataobj.GetTreeByPage(srvTag, dataobj.EndpointKeyNetwork)
			if err != nil {
				return ret, err
			}
			for _, n := range res.Networks {
				tagEndpoint := &dataobj.TagEndpoint{
					Ip:       n.ManageIp,
					HostName: n.Name,
					EnvCode:  n.EnvCode,
				}
				ret = append(ret, tagEndpoint)
			}
		}
	}
	for _, b := range bs {
		var te dataobj.TagEndpoint
		err := json.Unmarshal(b, &te)
		if err != nil {
			return ret, err
		}
		ret = append(ret, &te)
	}
	return ret, nil
}

func SetEndpointForRedis(key string, endpoints []*dataobj.TagEndpoint) error {
	batch := 10
	size := len(endpoints)
	n := size/batch + 1
	if size%batch == 0 {
		n = size / batch
	}
	for i := 0; i < n; i++ {
		tmp := []interface{}{}
		start := i * batch
		end := start + batch
		if i == n-1 && size%batch != 0 {
			end = start + size%batch
		}
		for _, e := range endpoints[start:end] {
			if e == nil {
				logger.Errorf("tag endpoint is nil")
				continue
			}
			data, err := json.Marshal(e)
			if err != nil {
				return err
			}
			tmp = append(tmp, data)
		}
		// 写redis
		redisc.SADD(key, tmp)
	}
	return nil
}

func SetEndpointLock() (bool, error) {
	return redisc.SETNX(dataobj.EndpointKeyForLock, "1", 60*60)
}

func SetEndpointUnLock() {
	if ret := redisc.GET(dataobj.EndpointKeyForLock); ret != 1 {
		logger.Info("%s key is not exists")
		return
	}

	if err := redisc.DelKey(dataobj.EndpointKeyForLock); err != nil {
		logger.Errorf("unlock EndpointKeyForLock key error %v", err)
	}
}
