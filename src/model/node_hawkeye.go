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

//根据服务树id获取子孙节点
func SrvTreeDescendants(nid int64) ([]*SrvTree, error) {
	url := config.Get().SrvTree.Addr + "/srv_tree/descendants"

	m := map[string]int64{
		"currentNodeId": nid,
	}

	var result SrvResult
	err := httplib.Post(url).JSONBodyQuiet(m).SetTimeout(3 * time.Second).ToJSON(&result)
	if err != nil {
		err = fmt.Errorf("request srvTree descendants fail: nid:%v, err:%v", nid, err)
		logger.Error(err)
		return nil, err
	}

	if result.Status != 200 {
		return nil, fmt.Errorf("%v srvtree descendants status error", nid)
	}

	return result.SrvTree, nil

}

type SrvResult struct {
	Message string     `json:"message"`
	SrvTree []*SrvTree `json:"result"`
	Status  int        `json:"status"`
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
