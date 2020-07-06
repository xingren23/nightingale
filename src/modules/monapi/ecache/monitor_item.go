package ecache

import (
	"github.com/didi/nightingale/src/modules/monapi/dataobj"
	"sync"
)

// 服务树节点id -> 节点串
type MonitorItemCacheMap struct {
	sync.RWMutex
	Data map[int64]*dataobj.MonitorItem
}

var MonitorItemCache *MonitorItemCacheMap

func NewMonitorItemCache() *MonitorItemCacheMap {
	return &MonitorItemCacheMap{
		Data: make(map[int64]*dataobj.MonitorItem),
	}
}

func (this *MonitorItemCacheMap) Get(key int64) (*dataobj.MonitorItem, bool) {
	this.RLock()
	defer this.RUnlock()

	value, exists := this.Data[key]
	return value, exists
}
