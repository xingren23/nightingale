package ecache

import (
	"sync"

	"github.com/didi/nightingale/src/modules/monapi/meicai"
)

type AppCacheList struct {
	sync.RWMutex
	Data []*meicai.App
}

func NewAppCache() *AppCacheList {
	return &AppCacheList{
		Data: []*meicai.App{},
	}
}

func (this *AppCacheList) SetAll(vals []*meicai.App) {
	this.Lock()
	defer this.Unlock()
	this.Data = vals
}

func (this *AppCacheList) Len() int {
	this.RLock()
	defer this.RUnlock()
	return len(this.Data)
}

func (this *AppCacheList) GetAll() []*meicai.App {
	this.RLock()
	defer this.RUnlock()
	var apps []*meicai.App
	for _, app := range this.Data {
		apps = append(apps, app)
	}
	return apps
}

type HostCacheList struct {
	sync.RWMutex
	Data []*meicai.CmdbHost
}

func NewHostCache() *HostCacheList {
	return &HostCacheList{
		Data: []*meicai.CmdbHost{},
	}
}

func (this *HostCacheList) SetAll(vals []*meicai.CmdbHost) {
	this.Lock()
	defer this.Unlock()
	this.Data = vals
}

func (this *HostCacheList) Len() int {
	this.RLock()
	defer this.RUnlock()
	return len(this.Data)
}

func (this *HostCacheList) GetAll() []*meicai.CmdbHost {
	this.RLock()
	defer this.RUnlock()
	var hosts []*meicai.CmdbHost
	for _, host := range this.Data {
		hosts = append(hosts, host)
	}
	return hosts
}

type InstanceCacheList struct {
	sync.RWMutex
	Data []*meicai.Instance
}

func NewInstanceCache() *InstanceCacheList {
	return &InstanceCacheList{
		Data: []*meicai.Instance{},
	}
}

func (this *InstanceCacheList) SetAll(vals []*meicai.Instance) {
	this.Lock()
	defer this.Unlock()
	this.Data = vals
}

func (this *InstanceCacheList) Len() int {
	this.RLock()
	defer this.RUnlock()
	return len(this.Data)
}

func (this *InstanceCacheList) GetAll() []*meicai.Instance {
	this.RLock()
	defer this.RUnlock()
	var instances []*meicai.Instance
	for _, instance := range this.Data {
		instances = append(instances, instance)
	}
	return instances
}

type NetworkCacheList struct {
	sync.RWMutex
	Data []*meicai.Network
}

func NewNetworkCache() *NetworkCacheList {
	return &NetworkCacheList{
		Data: []*meicai.Network{},
	}
}

func (this *NetworkCacheList) SetAll(vals []*meicai.Network) {
	this.Lock()
	defer this.Unlock()
	this.Data = vals
}

func (this *NetworkCacheList) Len() int {
	this.RLock()
	defer this.RUnlock()
	return len(this.Data)
}

func (this *NetworkCacheList) GetAll() []*meicai.Network {
	this.RLock()
	defer this.RUnlock()
	var networks []*meicai.Network
	for _, network := range this.Data {
		networks = append(networks, network)
	}
	return networks
}
