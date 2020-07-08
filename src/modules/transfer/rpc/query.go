package rpc

import (
	"github.com/didi/nightingale/src/dataobj"
	"github.com/didi/nightingale/src/modules/transfer/backend"
	"github.com/toolkits/pkg/logger"
)

func (t *Transfer) Query(args []dataobj.QueryData, reply *dataobj.QueryDataResp) error {
	storage, err := backend.GetStorageFor("")
	if err != nil {
		logger.Warningf("Could not find storage ")
		return err
	}
	reply.Data = storage.QueryData(args)
	return nil
}
