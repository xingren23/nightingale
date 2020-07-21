package ecache

import (
	"sync"

	"github.com/didi/nightingale/src/modules/monapi/meicai"
)

type IpInstancesCacheMap struct {
	sync.RWMutex
	Data map[string][]*meicai.Instance
}

var IpInstsCache *IpInstancesCacheMap

func NewIpInstsCache() *IpInstancesCacheMap {
	return &IpInstancesCacheMap{
		Data: map[string][]*meicai.Instance{},
	}
}

func (this *IpInstancesCacheMap) SetAll(vals map[string][]*meicai.Instance) {
	this.Lock()
	defer this.Unlock()
	this.Data = vals
}

func (this *IpInstancesCacheMap) GetByIp(ip string) ([]*meicai.Instance, bool) {
	this.RLock()
	defer this.RUnlock()
	value, exists := this.Data[ip]
	return value, exists
}

func (this *IpInstancesCacheMap) Len() int {
	this.RLock()
	defer this.RUnlock()
	return len(this.Data)
}
