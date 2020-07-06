package ecache

import (
	"github.com/didi/nightingale/src/modules/monapi/dataobj"
	"sync"
)

// 指标名称 -> 元数据信息
type MonitorItemCacheMap struct {
	sync.RWMutex
	Data map[string]*dataobj.MonitorItem
}

var MonitorItemCache *MonitorItemCacheMap

func NewMonitorItemCache() *MonitorItemCacheMap {
	return &MonitorItemCacheMap{
		Data: make(map[string]*dataobj.MonitorItem),
	}
}

func (this *MonitorItemCacheMap) Get(key string) (*dataobj.MonitorItem, bool) {
	this.RLock()
	defer this.RUnlock()

	value, exists := this.Data[key]
	return value, exists
}
