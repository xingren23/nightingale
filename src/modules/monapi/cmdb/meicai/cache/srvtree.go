package cache

import (
	"fmt"
	"sync"

	"github.com/didi/nightingale/src/modules/monapi/cmdb/dataobj"
)

// srvtree cache : id -> node , path -> node
type SrvTreeCache struct {
	sync.RWMutex
	idNodes   map[int64]*dataobj.Node
	pathNodes map[string]*dataobj.Node
}

func NewSrvTreeCache() *SrvTreeCache {
	return &SrvTreeCache{
		idNodes:   make(map[int64]*dataobj.Node),
		pathNodes: make(map[string]*dataobj.Node),
	}
}

func (cache *SrvTreeCache) GetById(key int64) (*dataobj.Node, bool) {
	cache.RLock()
	defer cache.RUnlock()
	value, exists := cache.idNodes[key]
	return value, exists
}

func (cache *SrvTreeCache) GetByIds(ids []int64) ([]*dataobj.Node, error) {
	cache.RLock()
	defer cache.RUnlock()

	ret := make([]*dataobj.Node, len(ids))
	for _, id := range ids {
		if node, exist := cache.idNodes[id]; exist {
			ret = append(ret, node)
		} else {
			return ret, fmt.Errorf("srvtree id not exist, %s", id)
		}
	}
	return ret, nil
}

func (cache *SrvTreeCache) GetByPath(path string) (*dataobj.Node, bool) {
	cache.RLock()
	defer cache.RUnlock()
	value, exists := cache.pathNodes[path]
	return value, exists
}

func (cache *SrvTreeCache) GetByPaths(paths []string) ([]*dataobj.Node, error) {
	cache.RLock()
	defer cache.RUnlock()

	ret := make([]*dataobj.Node, len(paths))
	for _, path := range paths {
		if node, exist := cache.pathNodes[path]; exist {
			ret = append(ret, node)
		} else {
			return ret, fmt.Errorf("srvtree path not exist, %s", path)
		}
	}
	return ret, nil
}

func (cache *SrvTreeCache) SetAll(nodes []*dataobj.Node) {
	cache.Lock()
	defer cache.Unlock()

	idNodes := make(map[int64]*dataobj.Node, len(nodes))
	pathNodes := make(map[string]*dataobj.Node, len(nodes))
	for _, node := range nodes {
		idNodes[node.Id] = node
		pathNodes[node.Path] = node
	}
	cache.idNodes = idNodes
	cache.pathNodes = pathNodes
}

func (cache *SrvTreeCache) Len() int {
	cache.RLock()
	defer cache.RUnlock()
	return len(cache.idNodes)
}

func (cache *SrvTreeCache) GetAll() []*dataobj.Node {
	cache.RLock()
	defer cache.RUnlock()

	ret := make([]*dataobj.Node, len(cache.idNodes))
	for _, node := range cache.idNodes {
		ret = append(ret, node)
	}
	return ret
}
