package ecache

import (
	"sync"

	"github.com/didi/nightingale/src/modules/monapi/dataobj"
)

type AppCacheList struct {
	sync.RWMutex
	Data []*dataobj.App
}

var AppCache *AppCacheList

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
