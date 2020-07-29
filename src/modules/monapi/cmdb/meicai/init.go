package meicai

import (
	"fmt"
	"time"

	"github.com/didi/nightingale/src/dataobj"
	cmdbdataobj "github.com/didi/nightingale/src/modules/monapi/cmdb/dataobj"
	"github.com/didi/nightingale/src/modules/monapi/cmdb/meicai/cache"
	"github.com/didi/nightingale/src/modules/monapi/config"
	"github.com/toolkits/pkg/logger"
)

type MeicaiSection struct {
	Enabled bool   `yaml:"enabled"`
	Name    string `yaml:"name"`
	Timeout int    `yaml:"timeout"`
	OpsAddr string `yaml:"opsAddr"`
}

type Meicai struct {
	Timeout      int
	OpsAddr      string
	srvTreeCache *cache.SrvTreeCache
}

const (
	OpsSrvtreeRootPath  = "/srv_tree/tree"
	OpsApiResourcerPath = "/api/resource/query"
)

const (
	CmdbSourceInst = "instance"
	CmdbSourceApp  = "app"
	CmdbSourceNet  = "network"
	CmdbSourceHost = "host"
)

func (meicai *Meicai) Init() {
	meicai.srvTreeCache = cache.NewSrvTreeCache()

	// init srvtree
	meicai.InitNode()

	// init resource

}

func (meicai *Meicai) InitSrvTagEndpoint() error {
	start := time.Now()
	// 加锁
	for {
		ok, err := cache.SetEndpointLock()
		if ok {
			break
		}
		logger.Warningf("endpoint lock is exists or error %v,sleep 2s", err)
		time.Sleep(2 * time.Second)
	}
	defer cache.SetEndpointUnLock()

	keys, err := cache.ScanRedisEndpointKeys()
	if err != nil {
		return err
	}
	logger.Infof("redis cache size %d", len(keys))
	if len(keys) > 0 {
		return nil
	}

	// 遍历节点
	logger.Info("start init srvTag_endpoint.")
	nodes := meicai.srvTreeCache.GetAll()
	for _, node := range nodes {
		nodeStr := node.Path
		url := fmt.Sprintf("%s%s", meicai.OpsAddr, OpsApiResourcerPath)

		initNodeHosts(url, meicai.Timeout, nodeStr)
		initNodeNetworks(url, meicai.Timeout, nodeStr)

	}
	logger.Infof("init srvTag_endpoints redis cache elapsed %s ms", time.Since(start))
	return nil
}

func initNodeHosts(url string, timeout int, nodeStr string) error {
	// 主机资源
	endpoints, err := EndpointUnderNodeGets(url, timeout, nodeStr, CmdbSourceHost)
	if err != nil {
		return err
	}
	pms, dockers := splitHosts(endpoints)
	pmKey := cache.RedisSrvTagKey(config.EndpointKeyPM, nodeStr)
	if err = cache.SetEndpointForRedis(pmKey, pms); err != nil {
		return err
	}
	dockerKey := cache.RedisSrvTagKey(config.EndpointKeyDocker, nodeStr)
	if err = cache.SetEndpointForRedis(dockerKey, dockers); err != nil {
		return err
	}
	return nil
}

func initNodeNetworks(url string, timeout int, nodeStr string) error {
	endpoints, err := EndpointUnderNodeGets(url, timeout, nodeStr, CmdbSourceNet)
	if err != nil {
		return err
	}
	netKey := cache.RedisSrvTagKey(config.EndpointKeyNetwork, nodeStr)
	if err = cache.SetEndpointForRedis(netKey, endpoints); err != nil {
		return err
	}
	return nil
}

func splitHosts(endpoints []*cmdbdataobj.Endpoint) (pms []*cmdbdataobj.Endpoint, dockers []*cmdbdataobj.Endpoint) {
	for _, endpoint := range endpoints {
		tags, err := dataobj.SplitTagsString(endpoint.Tags)
		if err != nil {
			continue
		}
		if tags["type"] == "DOCKER" {
			dockers = append(dockers, endpoint)
		} else {
			pms = append(pms, endpoint)
		}
	}
}
