package ecache

import (
	"sync"

	"github.com/didi/nightingale/src/model"
)

// 指标名称 -> 元数据信息
type MonitorItemCacheMap struct {
	sync.RWMutex
	Data map[string]*model.MonitorItem
}

var MonitorItemCache *MonitorItemCacheMap

func NewMonitorItemCache() *MonitorItemCacheMap {
	return &MonitorItemCacheMap{
		Data: make(map[string]*model.MonitorItem),
	}
}

func (this *MonitorItemCacheMap) Get(key string) (*model.MonitorItem, bool) {
	this.RLock()
	defer this.RUnlock()

	value, exists := this.Data[key]
	return value, exists
}

func (this *MonitorItemCacheMap) SetAll(items map[string]*model.MonitorItem) {
	this.Lock()
	defer this.Unlock()

	this.Data = items
}

func (this *MonitorItemCacheMap) GetAll() map[string]*model.MonitorItem {
	this.RLock()
	defer this.RUnlock()
	var monitorItemMap map[string]*model.MonitorItem
	for _, monitorItem := range this.Data {
		monitorItemMap[monitorItem.Metric] = monitorItem
	}
	return monitorItemMap
}
