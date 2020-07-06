package ecache

import (
	"fmt"
	"sync"
)

// 服务树节点串 + 资源类型 -> endpoint列表
type SrvTagEndpointCacheMap struct {
	sync.RWMutex
	Data map[string][]string
}

var SrvTagEndpointCache *SrvTagEndpointCacheMap

func NewSrvTagEndpointCache() *SrvTagEndpointCacheMap {
	return &SrvTagEndpointCacheMap{
		Data: make(map[string][]string),
	}
}

func (this *SrvTagEndpointCacheMap) Get(srvTag string, srvType string) ([]string, bool) {
	this.RLock()
	defer this.RUnlock()

	key := fmt.Sprintf("%s/%s", srvTag, srvType)
	value, exists := this.Data[key]
	return value, exists
}
