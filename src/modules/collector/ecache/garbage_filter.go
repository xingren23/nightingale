package ecache

import (
	"github.com/didi/nightingale/src/model"
	"strings"
	"sync"
)

var GarbageFilterCache *GarbageFilterCacheList

type GarbageFilterCacheList struct {
	sync.RWMutex
	Data []string
}

func NewGarbageFilterCache() *GarbageFilterCacheList {
	pc := GarbageFilterCacheList{
		Data: make([]string, 0),
	}
	return &pc
}

func (sc *GarbageFilterCacheList) SetAll(vals []model.ConfigInfo) {
	vs := []string{}
	for _, val := range vals {
		for _, v := range strings.Split(val.CfgValue, ",") {
			vs = append(vs, v)
		}
	}
	sc.Lock()
	defer sc.Unlock()
	sc.Data = vs
}

func (sc *GarbageFilterCacheList) Get() []string {
	sc.RLock()
	defer sc.RUnlock()
	return sc.Data
}