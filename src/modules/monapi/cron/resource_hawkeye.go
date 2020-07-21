package cron

import (
	"time"

	"github.com/didi/nightingale/src/modules/monapi/meicai"

	"github.com/didi/nightingale/src/modules/monapi/ecache"
	"github.com/didi/nightingale/src/toolkits/stats"
	"github.com/toolkits/pkg/logger"
)

func SyncResourceLoop() {
	// TODO : sync interval config
	duration := time.Second * time.Duration(60)
	for {
		time.Sleep(duration)
		logger.Debug("sync cmdb resource begin")
		err := SyncResource()
		if err != nil {
			stats.Counter.Set("cmdb_resource.sync.err", 1)
			logger.Error("sync cmdb resource fail: ", err)
		} else {
			logger.Debug("sync cmdb resource succ")
		}
	}
}

func SyncResource() error {
	start := time.Now()
	//应用
	if apps, err := meicai.GetAllApps(); err == nil {
		ecache.AppCache.SetAll(apps)
		logger.Infof("cache cmdb application size %d.", ecache.AppCache.Len())
	}
	// 主机
	if hosts, err := meicai.GetAllHosts(); err == nil {
		ecache.HostCache.SetAll(hosts)
		logger.Infof("cache cmdb host size %d.", ecache.HostCache.Len())
	}
	// 实例
	if insts, err := meicai.GetAllInstances(); err == nil {
		ecache.InstanceCache.SetAll(insts)
		logger.Infof("cache cmdb instance size %d.", ecache.InstanceCache.Len())
	}
	// 网络
	if nets, err := meicai.GetAllNetworks(); err == nil {
		ecache.NetworkCache.SetAll(nets)
		logger.Infof("cache cmdb network size %d.", ecache.NetworkCache.Len())
	}
	logger.Infof("sync cmdb resource cache elapsed %s ms", time.Since(start))
	return nil
}
