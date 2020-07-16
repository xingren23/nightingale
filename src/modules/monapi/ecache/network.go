package ecache

import (
	"sync"

	"github.com/didi/nightingale/src/modules/monapi/dataobj"
)

type NetworkCacheList struct {
	sync.RWMutex
	Data []*dataobj.Network
}

var NetworkCache *NetworkCacheList

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
