package meicai

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"

	"github.com/didi/nightingale/src/modules/monapi/config"
	"github.com/toolkits/pkg/logger"
)

type SrvTreeNode struct {
	Id       int64          `json:"id"`
	ParentId int64          `json:"parentId"`
	Name     string         `json:"name"`
	NodeCode string         `json:"nodeCode"`
	Type     string         `json:"type"`
	TagStr   string         `json:"tagStr"`
	HasLeaf  bool           `json:"hasLeaf"`
	Children []*SrvTreeNode `json:"children"`
}

type SrvTreeNodesResult struct {
	Message string         `json:"message"`
	Result  []*SrvTreeNode `json:"result"`
	Status  int            `json:"status"`
}

type SrvTreeNodeResult struct {
	Message string       `json:"message"`
	Result  *SrvTreeNode `json:"result"`
	Status  int          `json:"status"`
}

type NetworkPageResult struct {
	Pagination Pagination `json:"pagination"`
	Networks   []*Network `json:"result"`
}

type NetworkResult struct {
	Message string             `json:"message"`
	Result  *NetworkPageResult `json:"result"`
}

type CommonResult struct {
	Apps     []*App      `json:"result"`
	Networks []*Network  `json:"result"`
	Insts    []*Instance `json:"result"`
	Hosts    []*CmdbHost `json:"result"`
}

// 获取整棵服务树
func GetSrvTree() ([]*SrvTreeNode, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	url := fmt.Sprintf("%s%s", config.Get().Api.OpsAddr, config.OpsSrvtreeRootPath)
	data, err := RequestByPost(url, map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	var m SrvTreeNodesResult
	err = json.Unmarshal(data, &m)
	if err != nil {
		logger.Errorf("Cache Instance Parse JSON %v.", err)
		return nil, err
	}
	return m.Result, nil
}

// 获取服务树节点信息
func GetTreeById(nid int64) (*SrvTreeNode, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	url := fmt.Sprintf("%s%s/%d", config.Get().Api.OpsAddr, config.OpsSrvtreePath, nid)
	data, err := RequestByGet(url)
	if err != nil {
		return nil, err
	}

	var m SrvTreeNodeResult
	err = json.Unmarshal(data, &m)
	if err != nil {
		logger.Errorf("Cache Instance Parse JSON %v.", err)
		return nil, err
	}
	return m.Result, nil
}

// 获取服务树资源
func GetTreeResources(expr, cmdbSourceType string) (*CommonResult, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	commonRet := &CommonResult{
		Apps:     []*App{},
		Networks: []*Network{},
		Insts:    []*Instance{},
		Hosts:    []*CmdbHost{},
	}
	url := fmt.Sprintf("%s%s", config.Get().Api.OpsAddr, config.OpsApiResourcerPath)
	params := make(map[string]interface{})
	page := Pagination{
		PageNo:    1,
		PageSize:  100,
		TotalPage: 999,
	}
	params["sourceType"] = cmdbSourceType
	params["expression"] = expr

	for page.PageNo <= page.TotalPage {
		params["pagination"] = page

		data, err := RequestByPost(url, params)
		if err != nil {
			logger.Printf("%s", params)
			return nil, err
		}

		var pageRet Pagination
		switch cmdbSourceType {
		case config.CmdbSourceHost:
			var res CmdbHostResult
			err = json.Unmarshal(data, &res)
			if err != nil {
				logger.Errorf("Error: Parse CmdbHostResult JSON %v.", err)
				return nil, err
			}
			if res.Result == nil {
				logger.Error("Error: get tree resource host result nil")
				return nil, err
			}
			pageRet = res.Result.Pagination
			commonRet.Hosts = append(commonRet.Hosts, res.Result.Hosts...)
		case config.CmdbSourceNet:
			var res NetworkResult
			err = json.Unmarshal(data, &res)
			if err != nil {
				logger.Errorf("Error: Parse NetworkResult JSON %v.", err)
				return nil, err
			}
			if res.Result == nil {
				logger.Error("Error: get tree resource network result nil")
				return nil, err
			}
			pageRet = res.Result.Pagination
			commonRet.Networks = append(commonRet.Networks, res.Result.Networks...)
		}

		if pageRet.PageNo == 0 {
			return commonRet, fmt.Errorf("page result is nil")
		}
		page.PageNo++
		page.TotalPage = pageRet.TotalPage
	}

	return commonRet, nil
}
