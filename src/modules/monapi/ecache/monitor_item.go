package ecache

import (
	"github.com/didi/nightingale/src/model"
	"sync"
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
