package ecache

import (
	"sync"

	"github.com/didi/nightingale/src/modules/monapi/meicai"
)

// 指标名称 -> 元数据信息
type MonitorItemCacheMap struct {
	sync.RWMutex
	Data map[string]*meicai.MonitorItem
}

var MonitorItemCache *MonitorItemCacheMap

func NewMonitorItemCache() *MonitorItemCacheMap {
	return &MonitorItemCacheMap{
		Data: make(map[string]*meicai.MonitorItem),
	}
}

func (this *MonitorItemCacheMap) Get(key string) (*meicai.MonitorItem, bool) {
	this.RLock()
	defer this.RUnlock()

	value, exists := this.Data[key]
	return value, exists
}

func (this *MonitorItemCacheMap) SetAll(items map[string]*meicai.MonitorItem) {
	this.Lock()
	defer this.Unlock()

	this.Data = items
}

func (this *MonitorItemCacheMap) GetAll() map[string]*meicai.MonitorItem {
	this.RLock()
	defer this.RUnlock()
	var monitorItemMap map[string]*meicai.MonitorItem
	for _, monitorItem := range this.Data {
		monitorItemMap[monitorItem.Metric] = monitorItem
	}
	return monitorItemMap
}
