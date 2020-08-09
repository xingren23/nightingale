package cache

import (
	"strings"
	"sync"

	"github.com/didi/nightingale/src/model"
)

var GarbageCache *GarbageCacheList

type GarbageCacheList struct {
	sync.RWMutex
	Data []string
}

func NewGarbageCache() *GarbageCacheList {
	pc := GarbageCacheList{
		Data: make([]string, 0),
	}
	return &pc
}

func (sc *GarbageCacheList) SetAll(vals []model.ConfigInfo) {
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

func (sc *GarbageCacheList) Get() []string {
	sc.RLock()
	defer sc.RUnlock()
	return sc.Data
}
