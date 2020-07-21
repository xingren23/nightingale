package ecache

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/didi/nightingale/src/modules/monapi/meicai"

	"github.com/didi/nightingale/src/model"
	"github.com/didi/nightingale/src/modules/monapi/dataobj"
	"github.com/didi/nightingale/src/toolkits/address"
	"github.com/toolkits/pkg/logger"
	"github.com/toolkits/pkg/net/httplib"
)

var Resource ResourceSection

func Init(res ResourceSection) {
	Resource = res
	// fixme : 缓存构建失败 collector 能不能正常启动 ？
	AppCache = NewAppCache()
	HostCache = NewHostCache()
	InstanceCache = NewInstanceCache()
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
	// fixme : err 输出日志
	appResp, err := getApps()
	if err != nil {
		return err
	}
	hostResp, err := getHost()
	if err != nil {
		return err
	}
	networkResp, err := getNetwork()
	if err != nil {
		return err
	}
	instanceResp, err := getInstance()
	if err != nil {
		return err
	}
	monitorItemResp, err := getMonitorItem()
	if err != nil {
		return err
	}
	garbageFilterResp, err := getGarbageFiltersRetry()
	if err != nil {
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
	for _, instance := range instanceResp.Dat {
		instanceMap[instance.UUID] = instance
	}
	InstanceCache.SetAll(instanceMap)

	monitorItemMap := make(map[string]*meicai.MonitorItem)
	for _, monitorItem := range monitorItemResp.Dat {
		monitorItemMap[monitorItem.Metric] = monitorItem
	}
	MonitorItemCache.SetAll(monitorItemMap)

	return nil

}

type ResourceSection struct {
	AppApi           string `yaml:"appApi"`
	InstanceApi      string `yaml:"instanceApi"`
	NetworkApi       string `yaml:"networkApi"`
	HostApi          string `yaml:"hostApi"`
	MonitorItemApi   string `yaml:"monitorItemApi"`
	GarbageFilterApi string `yaml:"garbageFilterApi"`
	Timeout          int    `yaml:"timeout"`
}

func getApps() (AppResp, error) {
	addrs := address.GetHTTPAddresses("monapi")
	i := rand.Intn(len(addrs))
	addr := addrs[i]

	var res AppResp
	var err error

	url := fmt.Sprintf("http://%s%s", addr, Resource.AppApi)
	err = httplib.Get(url).SetTimeout(time.Duration(Resource.Timeout) * time.Millisecond).ToJSON(&res)
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

	url := fmt.Sprintf("http://%s%s", addr, Resource.HostApi)
	err = httplib.Get(url).SetTimeout(time.Duration(Resource.Timeout) * time.Millisecond).ToJSON(&res)
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

	url := fmt.Sprintf("http://%s%s", addr, Resource.InstanceApi)
	err = httplib.Get(url).SetTimeout(time.Duration(Resource.Timeout) * time.Millisecond).ToJSON(&res)
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

	url := fmt.Sprintf("http://%s%s", addr, Resource.NetworkApi)
	err = httplib.Get(url).SetTimeout(time.Duration(Resource.Timeout) * time.Millisecond).ToJSON(&res)
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

	url := fmt.Sprintf("http://%s%s", addr, Resource.MonitorItemApi)
	err = httplib.Get(url).SetTimeout(time.Duration(Resource.Timeout) * time.Millisecond).ToJSON(&res)
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
	Dat map[string]*meicai.MonitorItem `json:"dat"`
	Err string                         `json:"err"`
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

	url := fmt.Sprintf("http://%s%s", addr, Resource.GarbageFilterApi)
	err = httplib.Get(url).SetTimeout(time.Duration(Resource.Timeout) * time.Millisecond).ToJSON(&res)
	if err != nil {
		err = fmt.Errorf("get GarbageFilter config from remote:%s failed, error:%v", url, err)
	}

	return res, err
}
