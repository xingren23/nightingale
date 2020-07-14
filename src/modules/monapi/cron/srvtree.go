package cron

import (
	"github.com/toolkits/pkg/logger"
	"time"

	"github.com/didi/nightingale/src/modules/monapi/dataobj"
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
	return nil
}

func getTreeNodes() (map[int64]string, error) {
	nodeMap := make(map[int64]string)
	// 获取服务树
	nodes, err := dataobj.GetSrvTree()
	if err != nil {
		return nodeMap, err
	}
	getNodes(nodes, nodeMap)
	return nodeMap, nil
}

func getNodes(nodes []*dataobj.SrvTreeNode, nodeMap map[int64]string) {
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
	keys, err := ecache.GetEndpointKeysFromRedis()
	if err != nil {
		return err
	}
	logger.Infof("srvTree size %d, redis cache size %d", len(nodeMap), len(keys))
	// redis缓存数量少于90%
	if len(keys) < len(nodeMap)/10*80 {
		// 加锁
		if ok, err := ecache.SetEndpointLock(); !ok || err != nil {
			logger.Infof("endpoint lock is exists or error %v", err)
			return nil
		}
		defer ecache.SetEndpointUnLock()
		logger.Info("start init srvTag_endpoint.")
		// 遍历节点
		for _, nodeStr := range nodeMap {
			// 主机资源
			res, err := dataobj.GetTreeByPage(nodeStr, dataobj.CmdbSourceHost)
			if err != nil {
				return err
			}
			pms := []*dataobj.TagEndpoint{}
			dockers := []*dataobj.TagEndpoint{}
			for _, host := range res.Hosts {
				e := &dataobj.TagEndpoint{
					Ip:       host.Ip,
					HostName: host.HostName,
					EnvCode:  host.EnvCode,
					Endpoint: host.Ip,
				}
				// TODO 切换接口，先用这个测试
				host, _ = ecache.HostCache.Get(host.Ip)
				if host.Type == "DOCKER" {
					dockers = append(dockers, e)
				} else {
					pms = append(pms, e)
				}
			}
			pmKey := dataobj.BuildKey(dataobj.EndpointKeyPM, nodeStr)
			if err = ecache.SetEndpointForRedis(pmKey, pms); err != nil {
				return err
			}
			dockerKey := dataobj.BuildKey(dataobj.EndpointKeyDocker, nodeStr)
			if err = ecache.SetEndpointForRedis(dockerKey, dockers); err != nil {
				return err
			}

			// 网络
			res, err = dataobj.GetTreeByPage(nodeStr, dataobj.CmdbSourceNet)
			if err != nil {
				return err
			}
			networks := []*dataobj.TagEndpoint{}
			for _, net := range res.Networks {
				e := &dataobj.TagEndpoint{
					Ip:       net.ManageIp,
					HostName: net.Name,
					EnvCode:  net.EnvCode,
					Endpoint: net.ManageIp,
				}
				networks = append(networks, e)
			}
			netKey := dataobj.BuildKey(dataobj.EndpointKeyNetwork, nodeStr)
			if err = ecache.SetEndpointForRedis(netKey, networks); err != nil {
				return err
			}
		}
	}
	logger.Infof("init srvTag_endpoints redis cache elapsed %s ms", time.Since(start))
	return nil
}
