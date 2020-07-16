package ecache

import (
	"sync"

	"github.com/didi/nightingale/src/modules/monapi/dataobj"
)

type InstanceCacheMap struct {
	sync.RWMutex
	Data map[string]*dataobj.Instance
}

var InstanceCache *InstanceCacheMap

func NewInstanceCache() *InstanceCacheMap {
	return &InstanceCacheMap{
		Data: map[string]*dataobj.Instance{},
	}
}

func (this *InstanceCacheMap) SetAll(vals map[string]*dataobj.Instance) {
	this.Lock()
	defer this.Unlock()
	this.Data = vals
}

func (this *InstanceCacheMap) GetByUUID(uuid string) (*dataobj.Instance, bool) {
	this.RLock()
	defer this.RUnlock()
	value, exists := this.Data[uuid]
	return value, exists
}

func (this *InstanceCacheMap) Len() int {
	this.RLock()
	defer this.RUnlock()
	return len(this.Data)
}
