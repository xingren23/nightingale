package cache

import (
	"sync"
)

type Endpoint struct {
	Id    int64  `json:"id"`
	Ident string `json:"ident"`
	Alias string `json:"alias"`
	Tags  string `json:"tags"`
}

var HostCache *HostCacheMap

func NewHostCache() *HostCacheMap {
	return &HostCacheMap{
		hostMap: map[string]*Endpoint{},
	}
}

// host/network-ident(ip) -> endpoint
type HostCacheMap struct {
	sync.RWMutex
	hostMap map[string]*Endpoint
}

func (cache *HostCacheMap) Get(ident string) (*Endpoint, bool) {
	cache.RLock()
	defer cache.RUnlock()
	value, exists := cache.hostMap[ident]
	return value, exists
}

func (this *HostCacheMap) SetAll(hosts map[string]*Endpoint) {
	this.Lock()
	defer this.Unlock()
	this.hostMap = hosts
}

var InstanceCache *InstanceCacheMap

func NewInstanceCache() *InstanceCacheMap {
	return &InstanceCacheMap{
		instanceMap: map[string]*Endpoint{},
	}
}

// instance(uuid) -> endpoint
type InstanceCacheMap struct {
	sync.RWMutex
	instanceMap map[string]*Endpoint
}

func (cache *InstanceCacheMap) Get(uuid string) (*Endpoint, bool) {
	cache.RLock()
	defer cache.RUnlock()
	value, exists := cache.instanceMap[uuid]
	return value, exists
}

func (this *InstanceCacheMap) SetAll(instances map[string]*Endpoint) {
	this.Lock()
	defer this.Unlock()
	this.instanceMap = instances
}
