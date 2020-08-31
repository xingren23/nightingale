package cache

import (
	"sync"

	"github.com/didi/nightingale/src/model"
)

// 指标名称 -> 元数据信息
type MetricInfoCacheMap struct {
	sync.RWMutex
	Data map[string]*model.MetricInfo
}

var MetricInfoCache *MetricInfoCacheMap

func NewMetricInfoCache() *MetricInfoCacheMap {
	return &MetricInfoCacheMap{
		Data: make(map[string]*model.MetricInfo),
	}
}

func (this *MetricInfoCacheMap) Get(key string) (*model.MetricInfo, bool) {
	this.RLock()
	defer this.RUnlock()

	value, exists := this.Data[key]
	return value, exists
}

func (this *MetricInfoCacheMap) SetAll(items map[string]*model.MetricInfo) {
	this.Lock()
	defer this.Unlock()

	this.Data = items
}

func (this *MetricInfoCacheMap) GetAll() map[string]*model.MetricInfo {
	this.RLock()
	defer this.RUnlock()
	var metricInfoMap map[string]*model.MetricInfo
	for _, metricInfo := range this.Data {
		metricInfoMap[metricInfo.Metric] = metricInfo
	}
	return metricInfoMap
}
