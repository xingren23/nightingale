package ecache

import (
	"sync"

	"github.com/didi/nightingale/src/modules/monapi/dataobj"
)

type InstanceCacheList struct {
	sync.RWMutex
	Data []*dataobj.Instance
}

var InstanceCache *InstanceCacheList

func NewInstanceCache() *InstanceCacheList {
	return &InstanceCacheList{
		Data: []*dataobj.Instance{},
	}
}

func (this *InstanceCacheList) SetAll(vals []*dataobj.Instance) {
	this.Lock()
	defer this.Unlock()
	this.Data = vals
}

func (this *InstanceCacheList) Len() int {
	this.RLock()
	defer this.RUnlock()
	return len(this.Data)
}

func (this *InstanceCacheList) GetAll() []*dataobj.Instance {
	this.RLock()
	defer this.RUnlock()
	var instances []*dataobj.Instance
	for _, instance := range this.Data {
		instances = append(instances, instance)
	}
	return instances
}
