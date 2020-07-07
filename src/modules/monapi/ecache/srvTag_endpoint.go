package ecache

import (
	"fmt"
	"sync"

	"github.com/didi/nightingale/src/modules/monapi/dataobj"
)

// 服务树节点串 + 资源类型 -> endpoint列表
type SrvTagEndpointCacheMap struct {
	sync.RWMutex
	Data map[string][]*dataobj.TagEndpoint
}

var SrvTagEndpointCache *SrvTagEndpointCacheMap

func NewSrvTagEndpointCache() *SrvTagEndpointCacheMap {
	return &SrvTagEndpointCacheMap{
		Data: make(map[string][]*dataobj.TagEndpoint),
	}
}

func (this *SrvTagEndpointCacheMap) GetByKey(srvTag string, srvType string) ([]*dataobj.TagEndpoint, bool) {
	this.RLock()
	defer this.RUnlock()

	key := fmt.Sprintf("%s/%s", srvTag, srvType)
	res := []*dataobj.TagEndpoint{}
	vals, exists := this.Data[key]
	if exists {
		for _, val := range vals {
			if val == nil {
				continue
			}
			res = append(res, val)
		}
	}
	return res, exists
}
