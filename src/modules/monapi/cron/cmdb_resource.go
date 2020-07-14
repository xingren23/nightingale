package cron

import (
	"github.com/didi/nightingale/src/modules/monapi/dataobj"
	"github.com/didi/nightingale/src/modules/monapi/ecache"
	"github.com/didi/nightingale/src/toolkits/stats"
	"github.com/toolkits/pkg/logger"
	"time"
)

func SyncCmdbResourceLoop() {
	duration := time.Second * time.Duration(60)
	for {
		time.Sleep(duration)
		logger.Debug("sync cmdb resource begin")
		err := SyncMaskconf()
		if err != nil {
			stats.Counter.Set("cmdb_resource.sync.err", 1)
			logger.Error("sync cmdb resource fail: ", err)
		} else {
			logger.Debug("sync cmdb resource succ")
		}
	}
}

func SyncCmdbResource() error {
	//应用
	if apps, err := dataobj.GetAppByPage(); err == nil {
		ecache.AppCache.SetAll(apps)
		logger.Infof("cache cmdb application size %d.", ecache.AppCache.Len())
	}
	// 主机
	if hosts, err := dataobj.GetHostByPage(); err == nil {
		ecache.HostCache.SetAll(hosts)
		logger.Infof("cache cmdb host size %d.", ecache.HostCache.Len())
	}
	return nil
}
