package cache

import (
	"github.com/didi/nightingale/src/model"
	"strings"
	"sync"
)

var SieveCache *SievesCache

func InitSieves() {
	SieveCache = NewSievesCache()
}

type SievesCache struct {
	sync.RWMutex
	Data []string
}

func NewSievesCache() *SievesCache {
	pc := SievesCache{
		Data: make([]string, 0),
	}
	return &pc
}

func (sc *SievesCache) SetAll(vals []model.ConfigInfo) {
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

func (sc *SievesCache) Get() []string {
	sc.RLock()
	defer sc.RUnlock()
	return sc.Data
}
