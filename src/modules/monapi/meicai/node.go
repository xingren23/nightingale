package meicai

import (
	"fmt"
	"strconv"
	"time"

	"github.com/didi/nightingale/src/model"
	"github.com/didi/nightingale/src/modules/monapi/config"
	tlogger "github.com/didi/nightingale/src/toolkits/logger"

	"github.com/toolkits/pkg/logger"
	"github.com/toolkits/pkg/net/httplib"
)

func GetNodeById(nid int64) (*model.Node, error) {
	url := fmt.Sprintf("%s%s/%s", config.Get().Api.OpsAddr, config.OpsSrvtreePath, strconv.FormatInt(nid, 10))

	var result SrvResultDetail
	err := httplib.Get(url).SetTimeout(3 * time.Second).ToJSON(&result)
	if err != nil {
		err = fmt.Errorf("request srvTree detail fail: nid:%v, err:%v", nid, err)
		logger.Error(err)
		return nil, err
	}

	if result.Status != 200 {
		err = fmt.Errorf("request srvTree detail status error: nid:%v, status:%v", nid, result.Status)
		logger.Error(err)
		return nil, err
	}

	if result.SrvTree == nil {
		return nil, fmt.Errorf("request srvTree detail is nil: nid:%v", nid)
	}

	tlogger.TimeoutWarning("GetNodeById", config.OpsSrvtreePath, time.Now(), float64(1))
	return &model.Node{
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
	url := config.Get().Api.OpsAddr + config.OpsSrvtreeDescendantsPath

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

	tlogger.TimeoutWarning("SrvTreeDescendants", config.OpsSrvtreeDescendantsPath, time.Now(), float64(1))
	return result.SrvTree, nil
}

//根据服务树id获取父节点
func SrvTreeAncestors(nid int64) ([]*SrvTree, error) {
	url := config.Get().Api.OpsAddr + config.OpsSrvtreeAncestorsPath

	m := map[string]int64{
		"currentNodeId": nid,
	}

	var result SrvResult
	err := httplib.Post(url).JSONBodyQuiet(m).SetTimeout(3 * time.Second).ToJSON(&result)
	if err != nil {
		err = fmt.Errorf("request srvTree ancestors fail: nid:%v, err:%v", nid, err)
		logger.Error(err)
		return nil, err
	}

	if result.Status != 200 {
		return nil, fmt.Errorf("%v srvtree Ancestors status error", nid)
	}

	tlogger.TimeoutWarning("SrvTreeAncestors", config.OpsSrvtreeDescendantsPath, time.Now(), float64(1))
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
