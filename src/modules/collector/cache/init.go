package cache

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/didi/nightingale/src/dataobj"

	"github.com/didi/nightingale/src/model"
	"github.com/didi/nightingale/src/toolkits/address"
	"github.com/toolkits/pkg/logger"
	"github.com/toolkits/pkg/net/httplib"
)

const (
	EndpointsApi   = "/api/portal/endpoints"
	InstancesApi   = "/api/portal/instances"
	MonitorItemApi = "/api/portal/monitor_item"
	GarbageApi     = "/api/portal/garbage"

	Timeout = 10000
)

func Init() {
	MetricHistory = NewHistory()
	ProcsCache = NewProcsCache()

	EndpointCache = NewEndpointCache()
	InstanceCache = NewInstanceCache()
	MonitorItemCache = NewMonitorItemCache()
	GarbageCache = NewGarbageCache()

	if err := syncResource(); err != nil {
		log.Fatalf("build resource cache fail: %v", err)
	}
	go loopSyncResource()
}

func loopSyncResource() {
	t1 := time.NewTicker(time.Duration(180) * time.Second)
	logger.Info("[cron] sync resource cache start...")
	for {
		<-t1.C
		syncResource()
	}
}

func syncResource() error {
	err := buildEndpointCache()
	err = buildInstanceCache()
	err = buildGarbageCache()
	err = buildMonitorItemCache()
	return err
}

// endpoints, retry monapi addr
func buildEndpointCache() error {
	endpointsResp, err := getEndpoints()
	if err != nil {
		logger.Error("build endpoints cache fail:", err)
		return err
	}
	hostMap := make(map[string]*Endpoint)
	for _, endpoint := range endpointsResp.Dat {
		tags, err := dataobj.SplitTagsString(endpoint.Tags)
		if err != nil {
			logger.Warningf("split tags %s failed, host % %s", endpoint.Tags, endpoint.Ident, err)
			continue
		}
		if value, ok := tags["type"]; ok {
			if value == "HOST" || value == "NETWORK" {
				hostMap[endpoint.Ident] = endpoint
			}
		}
	}
	EndpointCache.SetAll(hostMap)
	return nil
}

// instances, retry monapi addr
func buildInstanceCache() error {
	instancesResp, err := getInstances()
	if err != nil {
		logger.Error("build endpoints cache fail:", err)
		return err
	}
	instanceMap := make(map[string]*Instance)
	for _, instance := range instancesResp.Dat {
		tags, err := dataobj.SplitTagsString(instance.Tags)
		if err != nil {
			logger.Warningf("split tags %s failed, host % %s", instance.Tags, instance.Ident, err)
			continue
		}
		if uuid, ok := tags["uuid"]; ok {
			//基础服务排除 ( basic=true )
			if basic, ok := tags["basic"]; ok {
				flag, err := strconv.ParseBool(basic)
				if err == nil && flag {
					logger.Debugf("don't process basic app, %v", instance)
					continue
				}
			}
			instanceMap[uuid] = instance
		}
	}
	InstanceCache.SetAll(instanceMap)
	return nil
}

// 指标元数据, retry monapi addr
func buildMonitorItemCache() error {
	monitorItemResp, err := getMonitorItem()
	if err != nil {
		logger.Error("build monitorItem cache fail:", err)
		return err
	}
	monitorItemMap := make(map[string]*model.MonitorItem)
	for _, monitorItem := range monitorItemResp.Dat {
		monitorItemMap[monitorItem.Metric] = monitorItem
	}
	MonitorItemCache.SetAll(monitorItemMap)
	return nil
}

// 过滤指标, retry monapi addr
func buildGarbageCache() error {
	garbageFilterResp, err := getGarbageFilter()
	if err != nil {
		logger.Error("build garbageFilter cache fail:", err)
		return err
	}
	GarbageCache.SetAll(garbageFilterResp.Dat)
	return nil
}

type EndpointsResp struct {
	Dat []*Endpoint `json:"dat"`
	Err string      `json:"err"`
}

type InstancesResp struct {
	Dat []*Instance `json:"dat"`
	Err string      `json:"err"`
}

type MonitorItemResp struct {
	Dat map[string]*model.MonitorItem `json:"dat"`
	Err string                        `json:"err"`
}

type GarbageFilterResp struct {
	Dat []model.ConfigInfo `json:"dat"`
	Err string             `json:"err"`
}

func getEndpoints() (EndpointsResp, error) {
	var res EndpointsResp
	var err error

	addrs := address.GetHTTPAddresses("monapi")
	count := len(addrs)
	for _, i := range rand.Perm(count) {
		addr := addrs[i]
		url := fmt.Sprintf("http://%s%s", addr, EndpointsApi)
		err = httplib.Get(url).SetTimeout(time.Duration(Timeout) * time.Millisecond).ToJSON(&res)
		if err != nil {
			err = fmt.Errorf("get apps from remote:%s failed, error:%v", url, err)
			continue
		}
		if res.Dat == nil || len(res.Dat) == 0 {
			err = fmt.Errorf("get apps from remote:%s is nil, error:%v", url, err)
			continue
		}
		break
	}
	return res, err
}

func getInstances() (InstancesResp, error) {
	var res InstancesResp
	var err error

	addrs := address.GetHTTPAddresses("monapi")
	count := len(addrs)
	for _, i := range rand.Perm(count) {
		addr := addrs[i]
		url := fmt.Sprintf("http://%s%s", addr, InstancesApi)
		err = httplib.Get(url).SetTimeout(time.Duration(Timeout) * time.Millisecond).ToJSON(&res)
		if err != nil {
			err = fmt.Errorf("get apps from remote:%s failed, error:%v", url, err)
			continue
		}
		if res.Dat == nil || len(res.Dat) == 0 {
			err = fmt.Errorf("get apps from remote:%s is nil, error:%v", url, err)
			continue
		}
		break
	}
	return res, err
}

func getMonitorItem() (MonitorItemResp, error) {
	var res MonitorItemResp
	var err error

	addrs := address.GetHTTPAddresses("monapi")
	count := len(addrs)
	for _, i := range rand.Perm(count) {
		addr := addrs[i]
		url := fmt.Sprintf("http://%s%s", addr, MonitorItemApi)
		err = httplib.Get(url).SetTimeout(time.Duration(Timeout) * time.Millisecond).ToJSON(&res)
		if err != nil {
			err = fmt.Errorf("get monitorItem from remote:%s failed, error:%v", url, err)
		}
		if res.Dat == nil || len(res.Dat) == 0 {
			err = fmt.Errorf("get monitorItem from remote:%s is nil, error:%v", url, err)
		}
		break
	}
	return res, err
}

func getGarbageFilter() (GarbageFilterResp, error) {
	var res GarbageFilterResp
	var err error

	addrs := address.GetHTTPAddresses("monapi")
	count := len(addrs)
	for _, i := range rand.Perm(count) {
		addr := addrs[i]
		url := fmt.Sprintf("http://%s%s", addr, GarbageApi)
		err = httplib.Get(url).SetTimeout(time.Duration(Timeout) * time.Millisecond).ToJSON(&res)
		if err != nil {
			err = fmt.Errorf("get GarbageFilter config from remote:%s failed, error:%v", url, err)
			continue
		}
		break
	}
	return res, err
}
