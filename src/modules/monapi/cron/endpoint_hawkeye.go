package cron

import (
	"time"

	"github.com/didi/nightingale/src/model"
	"github.com/didi/nightingale/src/modules/monapi/ecache"
	"github.com/didi/nightingale/src/toolkits/stats"
	"github.com/toolkits/pkg/logger"
)

func SyncEndpointsLoop() {
	duration := time.Second * time.Duration(60)
	for {
		time.Sleep(duration)
		logger.Debug("sync endpoints begin")
		err := SyncEndpoints()
		if err != nil {
			stats.Counter.Set("endpoints.sync.err", 1)
			logger.Error("sync endpoints fail: ", err)
		} else {
			logger.Debug("sync endpoints succ")
		}
	}
}

func SyncEndpoints() error {
	start := time.Now()
	endpointMap := make(map[string]*model.Endpoint)
	keys, err := ecache.ScanRedisEndpointKeys()
	if err != nil {
		return err
	}
	for _, key := range keys {
		srvType, nodePath := ecache.SplitRedisKey(key)
		// 填充endpoint
		tagEndpoints, err := ecache.GetEndpointsFromRedis(srvType, nodePath)
		if err != nil {
			return err
		}
		for _, te := range tagEndpoints {
			endpointMap[te.Endpoint] = &model.Endpoint{
				Ident: te.Endpoint,
				Alias: te.HostName,
			}
		}
	}
	// 添加缓存
	ecache.EndpointCache.SetAll(endpointMap)
	logger.Infof("sync endpoints cache elapsed %s ms", time.Since(start))
	return nil
}
