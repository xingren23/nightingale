package cache

import "sync"

type AppInstance struct {
	Id     int64  `json:"id"`
	App    string `json:"app"`
	Env    string `json:"env"`
	Group  string `json:"group"`
	Ident  string `json:"ident"`
	Port   int    `json:"port"`
	Uuid   string `json:"uuid"`
	Tags   string `json:"tags"`
	NodeId int64  `json:"nodeId"`
}

var AppInstanceCache *AppInstanceCacheMap

func NewAppInstanceCache() *AppInstanceCacheMap {
	return &AppInstanceCacheMap{
		instanceMap: map[string]*AppInstance{},
	}
}

// instance(uuid) -> endpoint
type AppInstanceCacheMap struct {
	sync.RWMutex
	instanceMap map[string]*AppInstance
}

func (cache *AppInstanceCacheMap) Get(uuid string) (*AppInstance, bool) {
	cache.RLock()
	defer cache.RUnlock()
	value, exists := cache.instanceMap[uuid]
	return value, exists
}

func (this *AppInstanceCacheMap) SetAll(instances map[string]*AppInstance) {
	this.Lock()
	defer this.Unlock()
	this.instanceMap = instances
}
