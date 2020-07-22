package ecache

import (
	"sync"

	"github.com/didi/nightingale/src/dataobj"
)

type AppCacheList struct {
	sync.RWMutex
	Data []*dataobj.App
}

func NewAppCache() *AppCacheList {
	return &AppCacheList{
		Data: []*dataobj.App{},
	}
}

func (this *AppCacheList) SetAll(vals []*dataobj.App) {
	this.Lock()
	defer this.Unlock()
	this.Data = vals
}

func (this *AppCacheList) Len() int {
	this.RLock()
	defer this.RUnlock()
	return len(this.Data)
}

func (this *AppCacheList) GetAll() []*dataobj.App {
	this.RLock()
	defer this.RUnlock()
	var apps []*dataobj.App
	for _, app := range this.Data {
		apps = append(apps, app)
	}
	return apps
}

type HostCacheList struct {
	sync.RWMutex
	Data []*dataobj.CmdbHost
}

func NewHostCache() *HostCacheList {
	return &HostCacheList{
		Data: []*dataobj.CmdbHost{},
	}
}

func (this *HostCacheList) SetAll(vals []*dataobj.CmdbHost) {
	this.Lock()
	defer this.Unlock()
	this.Data = vals
}

func (this *HostCacheList) Len() int {
	this.RLock()
	defer this.RUnlock()
	return len(this.Data)
}

func (this *HostCacheList) GetAll() []*dataobj.CmdbHost {
	this.RLock()
	defer this.RUnlock()
	var hosts []*dataobj.CmdbHost
	for _, host := range this.Data {
		hosts = append(hosts, host)
	}
	return hosts
}

type InstanceCacheList struct {
	sync.RWMutex
	Data []*dataobj.Instance
}

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

type NetworkCacheList struct {
	sync.RWMutex
	Data []*dataobj.Network
}

func NewNetworkCache() *NetworkCacheList {
	return &NetworkCacheList{
		Data: []*dataobj.Network{},
	}
}

func (this *NetworkCacheList) SetAll(vals []*dataobj.Network) {
	this.Lock()
	defer this.Unlock()
	this.Data = vals
}

func (this *NetworkCacheList) Len() int {
	this.RLock()
	defer this.RUnlock()
	return len(this.Data)
}

func (this *NetworkCacheList) GetAll() []*dataobj.Network {
	this.RLock()
	defer this.RUnlock()
	var networks []*dataobj.Network
	for _, network := range this.Data {
		networks = append(networks, network)
	}
	return networks
}
