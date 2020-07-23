package falcon

import (
	"math/rand"
	"sync"
	"time"

	"github.com/toolkits/pkg/logger"

	"github.com/didi/nightingale/src/dataobj"
	"github.com/didi/nightingale/src/modules/collector/sys"
)

var (
	TransferLock    = new(sync.RWMutex)
	TransferClients = map[string]*SingleConnRpcClient{}
)

func SendMetrics(metrics []*dataobj.MetricValue, resp *TransferResponse) {
	addrs := sys.Config.FalconTransfer.Addrs
	for _, i := range rand.Perm(len(addrs)) {
		addr := addrs[i]
		if _, ok := TransferClients[addr]; !ok {
			initTransferClient(addr)
		}
		if updateMetrics(addr, metrics, resp) {
			break
		}
	}
}

func initTransferClient(addr string) {
	TransferLock.Lock()
	defer TransferLock.Unlock()
	if _, exists := TransferClients[addr]; !exists {
		TransferClients[addr] = &SingleConnRpcClient{
			RpcServer: addr,
			Timeout:   time.Duration(sys.Config.FalconTransfer.Timeout) * time.Millisecond,
		}
	}
}

func updateMetrics(addr string, metrics []*dataobj.MetricValue, resp *TransferResponse) bool {
	TransferLock.RLock()
	defer TransferLock.RUnlock()
	err := TransferClients[addr].Call("Transfer.Update", metrics, resp)
	if err != nil {
		logger.Errorf("call %s Transfer.Update fail, %s", addr, err)
		return false
	}
	return true
}
