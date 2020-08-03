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
	err := meicai.InitNode()
	if err != nil {
		logger.Errorf("init meicai node failed, %s", err)
		panic(err)
	}

	// init endpoint
	err = meicai.InitEndpoint()
	if err != nil {
		logger.Errorf("init meicai endpoint failed, %s", err)
		panic(err)
	}
}

func (meicai *Meicai) InitEndpoint() error {
	start := time.Now()
	// 检查缓存内容，有内容则不再初始化
	keys, err := cache.ScanRedisEndpointKeys()
	if err != nil {
		logger.Errorf("scan redis cache endpoints failed, %s", err)
		return err
	}
	if len(keys) > 0 {
		logger.Infof("redis cache inited, size %d", len(keys))
		return nil
	}
	logger.Infof("redis cache size %d", len(keys))

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

	// 遍历节点
	logger.Info("start init endpoint.")
	nodes := meicai.srvTreeCache.GetNodes()
	for _, node := range nodes {
		logger.Infof("init node endpoint, id=%d path=%s", node.Id, node.Path)
		nodeStr := node.Path
		url := fmt.Sprintf("%s%s", meicai.OpsAddr, OpsApiResourcerPath)

		initNodeHosts(url, meicai.Timeout, nodeStr)
		initNodeNetworks(url, meicai.Timeout, nodeStr)

	}
	logger.Infof("end init endpoints, elapsed %s ms", time.Since(start))
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
	return
}
