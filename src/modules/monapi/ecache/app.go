package ecache

import (
	"sync"

	"github.com/didi/nightingale/src/modules/monapi/dataobj"
)

// appId -> 应用信息
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

func (this *AppCacheMap) Get(key int64) (*dataobj.App, bool) {
	this.RLock()
	defer this.RUnlock()

	value, exists := this.Data[key]
	return value, exists
}

func (this *AppCacheMap) SetAll(vals []*dataobj.App) {
	if vals == nil || len(vals) == 0 {
		return
	}
	m := make(map[int64]*dataobj.App)
	for _, app := range vals {
		if app != nil {
			m[app.Id] = app
		}
	}
	this.Lock()
	defer this.Unlock()
	this.Data = m
}

func (this *AppCacheMap) Len() int {
	this.RLock()
	defer this.RUnlock()
	return len(this.Data)
}
