package cron

import (
	"fmt"
	"time"

	"github.com/didi/nightingale/src/model"

	"github.com/didi/nightingale/src/modules/monapi/mcache"

	"github.com/didi/nightingale/src/toolkits/stats"
	"github.com/toolkits/pkg/logger"
)

func SyncMetricInfoLoop() {
	duration := time.Second * time.Duration(60)
	for {
		time.Sleep(duration)
		logger.Debug("sync metric info begin")
		err := SyncMaskconf()
		if err != nil {
			stats.Counter.Set("metricinfo.sync.err", 1)
			logger.Error("sync metric info fail: ", err)
		} else {
			logger.Debug("sync metric info succ")
		}
	}
}

func SyncMetricInfo() error {
	items, err := model.MetricInfoAll()
	if err != nil {
		return fmt.Errorf("get metric info fail: %v", err)
	}

	m := make(map[string]*model.MetricInfo)
	size := len(items)
	for i := 0; i < size; i++ {
		m[items[i].Metric] = items[i]
	}

	mcache.MetricInfoCache.SetAll(m)
	return nil
}
