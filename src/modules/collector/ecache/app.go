package ecache

import (
	"sync"

	"github.com/didi/nightingale/src/modules/monapi/meicai"
)

type AppCacheMap struct {
	sync.RWMutex
	Data map[int64]*meicai.App
}

var AppCache *AppCacheMap

func NewAppCache() *AppCacheMap {
	return &AppCacheMap{
		Data: make(map[int64]*meicai.App),
	}
}

func (this *AppCacheMap) SetAll(vals map[int64]*meicai.App) {
	this.Lock()
	defer this.Unlock()
	this.Data = vals
}

func (this *AppCacheMap) GetById(id int64) (*meicai.App, bool) {
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
