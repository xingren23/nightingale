package ecache

import (
	"sync"

	"github.com/didi/nightingale/src/modules/monapi/dataobj"
)

type HostCacheList struct {
	sync.RWMutex
	Data []*dataobj.CmdbHost
}

var HostCache *HostCacheList

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
