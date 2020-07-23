package ecache

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/didi/nightingale/src/dataobj"

	"github.com/didi/nightingale/src/model"
	"github.com/didi/nightingale/src/toolkits/address"
	"github.com/toolkits/pkg/logger"
	"github.com/toolkits/pkg/net/httplib"
)

const (
	AppApi         = "/api/portal/app"
	InstanceApi    = "/api/portal/instance"
	NetworkApi     = "/api/portal/network"
	HostApi        = "/api/portal/host"
	MonitorItemApi = "/api/portal/monitor_item"
	GarbageApi     = "/api/portal/garbage"

	Timeout = 10000
)

func Init() {
	AppCache = NewAppCache()
	HostCache = NewHostCache()
	InstanceCache = NewInstanceCache()
	IpInstsCache = NewIpInstsCache()
	NetworkCache = NewNetworkCache()
	MonitorItemCache = NewMonitorItemCache()
	GarbageFilterCache = NewGarbageFilterCache()

	if err := syncResource(); err != nil {
		log.Fatalf("build resourceCache fail: %v", err)
	}
	go loopSyncResource()
}

func loopSyncResource() {
	t1 := time.NewTicker(time.Duration(180) * time.Second)

	logger.Info("[cron] sync resourceCache start...")
	for {
		<-t1.C
		syncResource()
	}
}

func syncResource() error {
	count := len(address.GetHTTPAddresses("monapi"))
	var err error
	for i := 0; i < count; i++ {
		err = buildResourceCache()
		if err != nil {
			continue
		}
	}
	return err
}

func buildResourceCache() error {
	appResp, err := getApps()
	if err != nil {
		logger.Error("build app cache fail:", err)
		return err
	}
	hostResp, err := getHost()
	if err != nil {
		logger.Error("build host cache fail:", err)
		return err
	}
	networkResp, err := getNetwork()
	if err != nil {
		logger.Error("build network cache fail:", err)
		return err
	}
	instanceResp, err := getInstance()
	if err != nil {
		logger.Error("build instance cache fail:", err)
		return err
	}
	monitorItemResp, err := getMonitorItem()
	if err != nil {
		logger.Error("build monitorItem cache fail:", err)
		return err
	}
	garbageFilterResp, err := getGarbageFiltersRetry()
	if err != nil {
		logger.Error("build garbageFilter cache fail:", err)
		return err
	}
	GarbageFilterCache.SetAll(garbageFilterResp)

	appMap := make(map[int64]*dataobj.App)
	for _, app := range appResp.Dat {
		appMap[app.Id] = app
	}
	AppCache.SetAll(appMap)

	hostMap := make(map[string]*dataobj.CmdbHost)
	for _, host := range hostResp.Dat {
		hostMap[host.Ip] = host
	}
	HostCache.SetAll(hostMap)

	networkMap := make(map[string]*dataobj.Network)
	for _, network := range networkResp.Dat {
		networkMap[network.ManageIp] = network
	}
	NetworkCache.SetAll(networkMap)

	instanceMap := make(map[string]*dataobj.Instance)
	ipInstsMap := make(map[string][]*dataobj.Instance)
	for _, instance := range instanceResp.Dat {
		instanceMap[instance.UUID] = instance
		app, ok := AppCache.GetById(instance.AppId)
		if !ok {
			continue
		}
		// 基础服务排除
		if app.Basic {
			continue
		}
		if _, ok := ipInstsMap[instance.IP]; !ok {
			ipInstsMap[instance.IP] = []*dataobj.Instance{}
		}
		ipInstsMap[instance.IP] = append(ipInstsMap[instance.IP], instance)
	}
	InstanceCache.SetAll(instanceMap)
	IpInstsCache.SetAll(ipInstsMap)

	monitorItemMap := make(map[string]*model.MonitorItem)
	for _, monitorItem := range monitorItemResp.Dat {
		monitorItemMap[monitorItem.Metric] = monitorItem
	}
	MonitorItemCache.SetAll(monitorItemMap)

	return nil

}

func getApps() (AppResp, error) {
	addrs := address.GetHTTPAddresses("monapi")
	i := rand.Intn(len(addrs))
	addr := addrs[i]

	var res AppResp
	var err error

	url := fmt.Sprintf("http://%s%s", addr, AppApi)
	err = httplib.Get(url).SetTimeout(time.Duration(Timeout) * time.Millisecond).ToJSON(&res)
	if err != nil {
		err = fmt.Errorf("get apps from remote:%s failed, error:%v", url, err)
	}

	if res.Dat == nil || len(res.Dat) == 0 {
		err = fmt.Errorf("get apps from remote:%s is nil, error:%v", url, err)
	}

	return res, err
}

func getHost() (HostResp, error) {
	addrs := address.GetHTTPAddresses("monapi")
	i := rand.Intn(len(addrs))
	addr := addrs[i]

	var res HostResp
	var err error

	url := fmt.Sprintf("http://%s%s", addr, HostApi)
	err = httplib.Get(url).SetTimeout(time.Duration(Timeout) * time.Millisecond).ToJSON(&res)
	if err != nil {
		err = fmt.Errorf("get host from remote:%s failed, error:%v", url, err)
	}

	if res.Dat == nil || len(res.Dat) == 0 {
		err = fmt.Errorf("get host from remote:%s is nil, error:%v", url, err)
	}

	return res, err
}

func getInstance() (InstanceResp, error) {
	addrs := address.GetHTTPAddresses("monapi")
	i := rand.Intn(len(addrs))
	addr := addrs[i]

	var res InstanceResp
	var err error

	url := fmt.Sprintf("http://%s%s", addr, InstanceApi)
	err = httplib.Get(url).SetTimeout(time.Duration(Timeout) * time.Millisecond).ToJSON(&res)
	if err != nil {
		err = fmt.Errorf("get instance from remote:%s failed, error:%v", url, err)
	}

	if res.Dat == nil || len(res.Dat) == 0 {
		err = fmt.Errorf("get instance from remote:%s is nil, error:%v", url, err)
	}

	return res, err
}

func getNetwork() (NetworkResp, error) {
	addrs := address.GetHTTPAddresses("monapi")
	i := rand.Intn(len(addrs))
	addr := addrs[i]

	var res NetworkResp
	var err error

	url := fmt.Sprintf("http://%s%s", addr, NetworkApi)
	err = httplib.Get(url).SetTimeout(time.Duration(Timeout) * time.Millisecond).ToJSON(&res)
	if err != nil {
		err = fmt.Errorf("get network from remote:%s failed, error:%v", url, err)
	}

	if res.Dat == nil || len(res.Dat) == 0 {
		err = fmt.Errorf("get network from remote:%s is nil, error:%v", url, err)
	}

	return res, err
}

func getMonitorItem() (MonitorItemResp, error) {
	addrs := address.GetHTTPAddresses("monapi")
	i := rand.Intn(len(addrs))
	addr := addrs[i]

	var res MonitorItemResp
	var err error

	url := fmt.Sprintf("http://%s%s", addr, MonitorItemApi)
	err = httplib.Get(url).SetTimeout(time.Duration(Timeout) * time.Millisecond).ToJSON(&res)
	if err != nil {
		err = fmt.Errorf("get monitorItem from remote:%s failed, error:%v", url, err)
	}

	if res.Dat == nil || len(res.Dat) == 0 {
		err = fmt.Errorf("get monitorItem from remote:%s is nil, error:%v", url, err)
	}

	return res, err
}

type AppResp struct {
	Dat []*dataobj.App `json:"dat"`
	Err string         `json:"err"`
}

type HostResp struct {
	Dat []*dataobj.CmdbHost `json:"dat"`
	Err string              `json:"err"`
}

type InstanceResp struct {
	Dat []*dataobj.Instance `json:"dat"`
	Err string              `json:"err"`
}

type NetworkResp struct {
	Dat []*dataobj.Network `json:"dat"`
	Err string             `json:"err"`
}
type MonitorItemResp struct {
	Dat map[string]*model.MonitorItem `json:"dat"`
	Err string                        `json:"err"`
}

type GarbageFilterResp struct {
	Dat []model.ConfigInfo `json:"dat"`
	Err string             `json:"err"`
}

func getGarbageFiltersRetry() ([]model.ConfigInfo, error) {
	count := len(address.GetHTTPAddresses("monapi"))
	var resp GarbageFilterResp
	var err error
	for i := 0; i < count; i++ {
		resp, err = getGarbageFilter()
		if err == nil {
			if resp.Err != "" {
				err = fmt.Errorf(resp.Err)
				continue
			}
			return resp.Dat, err
		}
	}

	return resp.Dat, err
}

func getGarbageFilter() (GarbageFilterResp, error) {
	addrs := address.GetHTTPAddresses("monapi")
	i := rand.Intn(len(addrs))
	addr := addrs[i]

	var res GarbageFilterResp
	var err error

	url := fmt.Sprintf("http://%s%s", addr, GarbageApi)
	err = httplib.Get(url).SetTimeout(time.Duration(Timeout) * time.Millisecond).ToJSON(&res)
	if err != nil {
		err = fmt.Errorf("get GarbageFilter config from remote:%s failed, error:%v", url, err)
	}

	return res, err
}
