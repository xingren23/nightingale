package model

import (
	"fmt"
	"github.com/didi/nightingale/src/modules/monapi/config"
	"github.com/toolkits/pkg/logger"
	"github.com/toolkits/pkg/net/httplib"
	"strconv"
	"time"
)

func GetNodeById(nid int64) (*Node, error) {
	url := config.Get().SrvTree.Addr + "/srv_tree/" + strconv.FormatInt(nid, 10)

	var result SrvResultDetail
	err := httplib.Get(url).SetTimeout(3 * time.Second).ToJSON(&result)
	if err != nil {
		err = fmt.Errorf("request srvTree detail fail: nid:%v, err:%v", nid, err)
		logger.Error(err)
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	if result.Status != 200 {
		err = fmt.Errorf("request srvTree detail status error: nid:%v, status:%v", nid, result.Status)
		logger.Error(err)
		return nil, err
	}

	return &Node{
		Id:   result.SrvTree.Id,
		Pid:  0,
		Name: result.SrvTree.Name,
		Path: result.SrvTree.NodePath,
		Leaf: 0,
		Note: result.SrvTree.NodeCode,
	}, err

}

type SrvTree struct {
	Code     string `json:"code"`
	Id       int64  `json:"id"`
	Name     string `json:"name"`
	NodeCode string `json:"nodeCode"`
	NodePath string `json:"tagStr"`
}

type SrvResultDetail struct {
	Message string   `json:"message"`
	SrvTree *SrvTree `json:"result"`
	Status  int      `json:"status"`
}
