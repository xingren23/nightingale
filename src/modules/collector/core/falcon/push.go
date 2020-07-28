package falcon

import (
	"fmt"
	"net/rpc"
	"sync"
	"time"

	"github.com/didi/nightingale/src/dataobj"
	"github.com/toolkits/net"
	"github.com/toolkits/pkg/logger"
)

type SingleConnRpcClient struct {
	sync.Mutex
	rpcClient *rpc.Client
	RpcServer string
	Timeout   time.Duration
}

func (this *SingleConnRpcClient) close() {
	if this.rpcClient != nil {
		this.rpcClient.Close()
		this.rpcClient = nil
	}
}

func (this *SingleConnRpcClient) insureConn() bool {
	var retry = 0
	var err error
	for {
		if this.rpcClient != nil {
			logger.Infof("rpc client is ready, %s", this.RpcServer)
			return true
		}

		this.rpcClient, err = net.JsonRpcClient("tcp", this.RpcServer, this.Timeout)
		if err == nil {
			logger.Infof("init rpc client success, %s", this.RpcServer)
			return true
		}

		retry++
		if retry > 3 {
			err = fmt.Errorf("conn to rpcServer %s failed, retry %d", this.RpcServer, retry)
			return false
		}
		logger.Errorf("dial %s fail retry %d: %v", this.RpcServer, retry, err)
		time.Sleep(this.Timeout / 3)
	}
}

func (this *SingleConnRpcClient) Call(method string, args interface{}, reply interface{}) error {

	this.Lock()
	defer this.Unlock()

	done := make(chan error)
	go func() {
		ok := this.insureConn()
		if ok {
			err := this.rpcClient.Call(method, args, reply)
			done <- err
		} else {
			done <- fmt.Errorf("insure conn failed, %s", this.RpcServer)
		}
	}()

	select {
	case <-time.After(this.Timeout):
		logger.Warningf("rpc call timeout %d milliseconds, %v => %v", this.Timeout.Milliseconds(), this.rpcClient,
			this.RpcServer)
		this.close()
	case err := <-done:
		if err != nil {
			this.close()
			return err
		}
	}

	return nil
}

// 直接发送到falcon-transfer( falcon内部处理 )
func Push(metrics []*dataobj.MetricValue) {
	if len(metrics) == 0 {
		return
	}

	logger.Debugf("=> <Total=%d> \n", len(metrics))
	logger.Debug("push falcon transfer item: ", metrics)

	var resp TransferResponse
	SendMetrics(metrics, &resp)

	logger.Debug("<=", &resp)
}
