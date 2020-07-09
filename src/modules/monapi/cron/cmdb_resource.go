package cron

import (
	"github.com/didi/nightingale/src/toolkits/stats"
	"github.com/toolkits/pkg/logger"
	"time"
)

func SyncCmdbResourceLoop() error {
	duration := time.Second * time.Duration(9)
	for {
		time.Sleep(duration)
		logger.Debug("sync maskconf begin")
		err := SyncMaskconf()
		if err != nil {
			stats.Counter.Set("maskconf.sync.err", 1)
			logger.Error("sync maskconf fail: ", err)
		} else {
			logger.Debug("sync maskconf succ")
		}
	}
}

func SyncCmdbResource() error {
	// 应用
	//if apps, err := dataobj.GetAppByPage(); err == nil {
	//	ecache.AppCache.SetAll(apps)
	//	logger.Infof("cache cmdb application size %d.", ecache.AppCache.Len())
	//}
	//// 主机
	//if _, err := dataobj.GetHostByPage(); err == nil {
	//
	//	logger.Infof("cache cmdb application size %d.", ecache.AppCache.Len())
	//}
	//return nil
}
