package ecache

import (
	"sync"

	"github.com/didi/nightingale/src/dataobj"
)

type AppCacheMap struct {
	sync.RWMutex
	Data map[int64]*dataobj.App
}

var AppCache *AppCacheMap

func NewAppCache() *AppCacheMap {
	return &AppCacheMap{
		Data: make(map[int64]*dataobj.App),
	}
}

func (this *AppCacheMap) SetAll(vals map[int64]*dataobj.App) {
	this.Lock()
	defer this.Unlock()
	this.Data = vals
}

func (this *AppCacheMap) GetById(id int64) (*dataobj.App, bool) {
	this.RLock()
	defer this.RUnlock()
	value, exists := this.Data[id]
	return value, exists
}

func (this *AppCacheMap) Len() int {
	this.RLock()
	defer this.RUnlock()
	return len(this.Data)
}

type HostCacheMap struct {
	sync.RWMutex
	Data map[string]*dataobj.CmdbHost
}

var HostCache *HostCacheMap

func NewHostCache() *HostCacheMap {
	return &HostCacheMap{
		Data: map[string]*dataobj.CmdbHost{},
	}
}

func (this *HostCacheMap) GetByIp(ip string) (*dataobj.CmdbHost, bool) {
	this.RLock()
	defer this.RUnlock()
	value, exists := this.Data[ip]
	return value, exists
}

func (this *HostCacheMap) SetAll(vals map[string]*dataobj.CmdbHost) {
	this.Lock()
	defer this.Unlock()
	this.Data = vals
}

func (this *HostCacheMap) Len() int {
	this.RLock()
	defer this.RUnlock()
	return len(this.Data)
}

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

type NetworkCacheMap struct {
	sync.RWMutex
	Data map[string]*dataobj.Network
}

var NetworkCache *NetworkCacheMap

func NewNetworkCache() *NetworkCacheMap {
	return &NetworkCacheMap{
		Data: map[string]*dataobj.Network{},
	}
}

func (this *NetworkCacheMap) SetAll(vals map[string]*dataobj.Network) {
	this.Lock()
	defer this.Unlock()
	this.Data = vals
}

func (this *NetworkCacheMap) GetByIp(ip string) (*dataobj.Network, bool) {
	this.RLock()
	defer this.RUnlock()
	value, exists := this.Data[ip]
	return value, exists
}

func (this *NetworkCacheMap) Len() int {
	this.RLock()
	defer this.RUnlock()
	return len(this.Data)
}
