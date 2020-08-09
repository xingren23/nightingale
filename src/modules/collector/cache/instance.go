package cache

import "sync"

type Instance struct {
	Id    int64  `json:"id"`
	Ident string `json:"ident"`
	Alias string `json:"alias"`
	Port  int    `json:"int"`
	Tags  string `json:"tags"`
}

var InstanceCache *InstanceCacheMap

func NewInstanceCache() *InstanceCacheMap {
	return &InstanceCacheMap{
		instanceMap: map[string]*Instance{},
	}
}

// instance(uuid) -> endpoint
type InstanceCacheMap struct {
	sync.RWMutex
	instanceMap map[string]*Instance
}

func (cache *InstanceCacheMap) Get(uuid string) (*Instance, bool) {
	cache.RLock()
	defer cache.RUnlock()
	value, exists := cache.instanceMap[uuid]
	return value, exists
}

func (this *InstanceCacheMap) SetAll(instances map[string]*Instance) {
	this.Lock()
	defer this.Unlock()
	this.instanceMap = instances
}
