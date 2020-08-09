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

var EndpointCache *EndpointCacheMap

func NewEndpointCache() *EndpointCacheMap {
	return &EndpointCacheMap{
		hostMap: map[string]*Endpoint{},
	}
}

// host/network-ident(ip) -> endpoint
type EndpointCacheMap struct {
	sync.RWMutex
	hostMap map[string]*Endpoint
}

func (cache *EndpointCacheMap) Get(ident string) (*Endpoint, bool) {
	cache.RLock()
	defer cache.RUnlock()
	value, exists := cache.hostMap[ident]
	return value, exists
}

func (this *EndpointCacheMap) SetAll(hosts map[string]*Endpoint) {
	this.Lock()
	defer this.Unlock()
	this.hostMap = hosts
}
