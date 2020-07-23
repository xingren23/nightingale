package cron

import (
	"time"

	"github.com/didi/nightingale/src/modules/monapi/config"

	"github.com/didi/nightingale/src/modules/monapi/meicai"

	"github.com/toolkits/pkg/logger"

	"github.com/didi/nightingale/src/modules/monapi/ecache"
	"github.com/didi/nightingale/src/toolkits/stats"
)

func SyncSrvTreeLoop() {
	duration := time.Second * time.Duration(180)
	for {
		time.Sleep(duration)
		logger.Debug("sync srvTree begin")
		err := SyncSrvTree()
		if err != nil {
			stats.Counter.Set("srvTree.sync.err", 1)
			logger.Error("sync srvTree fail: ", err)
		} else {
			logger.Debug("sync srvTree succ")
		}
	}
}

func SyncSrvTree() error {
	start := time.Now()
	nodeMap, err := getTreeNodes()
	if err != nil {
		return err
	}
	// 添加缓存
	ecache.SrvTreeCache.SetAll(nodeMap)
	// 初始化
	if err = InitSrvTagEndpoint(); err != nil {
		return err
	}
	logger.Infof("sync srvTree cache elapsed %s ms", time.Since(start))
	return nil
}

func getTreeNodes() (map[int64]string, error) {
	nodeMap := make(map[int64]string)
	// 获取服务树
	nodes, err := meicai.GetSrvTree()
	if err != nil {
		return nodeMap, err
	}
	getNodes(nodes, nodeMap)
	return nodeMap, nil
}

func getNodes(nodes []*meicai.SrvTreeNode, nodeMap map[int64]string) {
	if nodes == nil || len(nodes) == 0 {
		return
	}
	for _, n := range nodes {
		if n == nil {
			continue
		}
		// 排除buffer节点
		if n.NodeCode == "buffer" {
			continue
		}
		nodeMap[n.Id] = n.TagStr
		getNodes(n.Children, nodeMap)
	}
}

func InitSrvTagEndpoint() error {
	start := time.Now()
	// 比较缓存与服务树中节点数量
	nodeMap := ecache.SrvTreeCache.GetAll()
	for {
		// 加锁
		ok, err := ecache.SetEndpointLock()
		if ok {
			break
		}
		logger.Warningf("endpoint lock is exists or error %v,sleep 2s", err)
		time.Sleep(2 * time.Second)
	}
	defer ecache.SetEndpointUnLock()

	keys, err := ecache.ScanRedisEndpointKeys()
	if err != nil {
		return err
	}
	logger.Infof("srvTree size %d, redis cache size %d", len(nodeMap), len(keys))
	if len(keys) > 0 {
		return nil
	}
	logger.Info("start init srvTag_endpoint.")
	// 遍历节点
	for _, nodeStr := range nodeMap {
		// 主机资源
		res, err := meicai.GetTreeResources(nodeStr, config.CmdbSourceHost)
		if err != nil {
			return err
		}
		pms := []*ecache.TagEndpoint{}
		dockers := []*ecache.TagEndpoint{}
		for _, host := range res.Hosts {
			e := &ecache.TagEndpoint{
				Ip:       host.Ip,
				HostName: host.HostName,
				EnvCode:  host.EnvCode,
				Endpoint: host.Ip,
			}
			if host.Type == "DOCKER" {
				dockers = append(dockers, e)
			} else {
				pms = append(pms, e)
			}
		}
		pmKey := ecache.RedisSrvTagKey(config.EndpointKeyPM, nodeStr)
		if err = ecache.SetEndpointForRedis(pmKey, pms); err != nil {
			return err
		}
		dockerKey := ecache.RedisSrvTagKey(config.EndpointKeyDocker, nodeStr)
		if err = ecache.SetEndpointForRedis(dockerKey, dockers); err != nil {
			return err
		}

		// 网络
		res, err = meicai.GetTreeResources(nodeStr, config.CmdbSourceNet)
		if err != nil {
			return err
		}
		networks := []*ecache.TagEndpoint{}
		for _, net := range res.Networks {
			e := &ecache.TagEndpoint{
				Ip:       net.ManageIp,
				HostName: net.Name,
				EnvCode:  net.EnvCode,
				Endpoint: net.ManageIp,
			}
			networks = append(networks, e)
		}
		netKey := ecache.RedisSrvTagKey(config.EndpointKeyNetwork, nodeStr)
		if err = ecache.SetEndpointForRedis(netKey, networks); err != nil {
			return err
		}
	}
	logger.Infof("init srvTag_endpoints redis cache elapsed %s ms", time.Since(start))
	return nil
}