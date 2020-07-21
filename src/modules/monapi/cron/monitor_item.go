package cron

import (
	"fmt"
	"time"

	"github.com/didi/nightingale/src/modules/monapi/ecache"
	"github.com/didi/nightingale/src/modules/monapi/meicai"
	"github.com/didi/nightingale/src/toolkits/stats"
	"github.com/toolkits/pkg/logger"
)

func SyncMonitorItemLoop() {
	duration := time.Second * time.Duration(60)
	for {
		time.Sleep(duration)
		logger.Debug("sync monitorItem begin")
		err := SyncMaskconf()
		if err != nil {
			stats.Counter.Set("monitorItem.sync.err", 1)
			logger.Error("sync monitorItem fail: ", err)
		} else {
			logger.Debug("sync monitorItem succ")
		}
	}
}

func SyncMonitorItem() error {
	items, err := meicai.MonitorItemAll()
	if err != nil {
		return fmt.Errorf("get monitorItem fail: %v", err)
	}

	m := make(map[string]*meicai.MonitorItem)
	size := len(items)
	for i := 0; i < size; i++ {
		m[items[i].Metric] = items[i]
	}

	ecache.MonitorItemCache.SetAll(m)
	return nil
}
