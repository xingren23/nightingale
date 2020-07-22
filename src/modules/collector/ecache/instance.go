package ecache

import (
	"sync"

	"github.com/didi/nightingale/src/modules/monapi/meicai"
)

type InstanceCacheMap struct {
	sync.RWMutex
	Data map[string]*meicai.Instance
}

var InstanceCache *InstanceCacheMap

func NewInstanceCache() *InstanceCacheMap {
	return &InstanceCacheMap{
		Data: map[string]*meicai.Instance{},
	}
}

func (this *InstanceCacheMap) SetAll(vals map[string]*meicai.Instance) {
	this.Lock()
	defer this.Unlock()
	this.Data = vals
}

func (this *InstanceCacheMap) GetByUUID(uuid string) (*meicai.Instance, bool) {
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
