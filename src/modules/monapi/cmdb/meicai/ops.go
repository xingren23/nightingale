package meicai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/didi/nightingale/src/toolkits/str"

	"github.com/didi/nightingale/src/modules/monapi/config"

	"github.com/didi/nightingale/src/modules/monapi/cmdb/dataobj"
	jsoniter "github.com/json-iterator/go"
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

// 获取整棵服务树
func SrvTreeGets(url string, timeout int) ([]*dataobj.Node, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	data, err := RequestByPost(url, timeout, map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	var srvTreeNodesResult SrvTreeNodesResult
	err = json.Unmarshal(data, &srvTreeNodesResult)
	if err != nil {
		logger.Errorf("Parse JSON %v.", err)
		return nil, err
	}
	if srvTreeNodesResult.Status != 200 {
		logger.Errorf("ops request %s error, %s", url, srvTreeNodesResult.Message)
		return nil, fmt.Errorf(srvTreeNodesResult.Message)
	}

	return convert2Node(srvTreeNodesResult.Result)
}

func convert2Node(treeNodes []*SrvTreeNode) ([]*dataobj.Node, error) {
	if treeNodes == nil || len(treeNodes) == 0 {
		return make([]*dataobj.Node, 0), nil
	}

	nodes := make([]*dataobj.Node, 0)
	for _, srvNode := range treeNodes {
		node := &dataobj.Node{
			Id:   srvNode.Id,
			Pid:  srvNode.ParentId,
			Name: srvNode.NodeCode,
			Path: srvNode.TagStr,
			Note: srvNode.Name,
		}
		if srvNode.HasLeaf && len(srvNode.Children) == 0 {
			node.Leaf = 1
		}
		nodes = append(nodes, node)

		childNodes, err := convert2Node(srvNode.Children)
		if err != nil {
			logger.Errorf("convert2Node error, %v", err)
		}
		nodes = append(nodes, childNodes...)
	}
	return nodes, nil
}

type Pagination struct {
	PageNo      int `json:"pageNo"`
	PageSize    int `json:"pageSize"`
	Start       int `json:"start"`
	TotalPage   int `json:"totalPage"`
	TotalRecord int `json:"totalRecord"`
}

type App struct {
	Id        int64  `json:"id"`
	Code      string `json:"code"`
	Language  string `json:"language"`
	Name      string `json:"name"`
	SrvTreeId int64  `json:"srvTreeId"`
	Basic     bool   `json:"basic"`
}

type CmdbAppPageResult struct {
	Pagination Pagination `json:"pagination"`
	Hosts      []*App     `json:"result"`
}

type CmdbAppResult struct {
	Message string             `json:"message"`
	Result  *CmdbAppPageResult `json:"result"`
}

type CmdbHost struct {
	Id             int64  `json:"id"`
	Ip             string `json:"innerIp"`
	HostName       string `json:"hostName"`
	Type           string `json:"type"`
	EnvCode        string `json:"envCode"`
	DataCenterCode string `json:"dataCenterCode"`
}

type CmdbHostPageResult struct {
	Pagination Pagination  `json:"pagination"`
	Hosts      []*CmdbHost `json:"result"`
}

type CmdbHostResult struct {
	Message string              `json:"message"`
	Result  *CmdbHostPageResult `json:"result"`
}

type Instance struct {
	Id             int64  `json:"id"`
	AppCode        string `json:"appCode"`
	AppId          int64  `json:"appId"`
	DataCenterCode string `json:"dataCenterCode"`
	EnvCode        string `json:"envCode"`
	GroupCode      string `json:"groupCode"`
	IP             string `json:"ip"`
	HostName       string `json:"hostName"`
	Port           int    `json:"port"`
	UUID           string `json:"uuid"`
}

type CmdbInstancePageResult struct {
	Pagination Pagination  `json:"pagination"`
	Instances  []*Instance `json:"result"`
}

type CmdbInstanceResult struct {
	Message string                  `json:"message"`
	Result  *CmdbInstancePageResult `json:"result"`
}

type Network struct {
	Id             int64  `json:"id"`
	ManageIp       string `json:"manageIp"`
	Name           string `json:"name"`
	DataCenterCode string `json:"dataCenterCode"`
	EnvCode        string `json:"envCode"`
	Type           string `json:"type"`
	SerialNo       string `json:"serialNo"`
}

type CmdbNetworkPageResult struct {
	Pagination Pagination `json:"pagination"`
	Networks   []*Network `json:"result"`
}

type CmdbNetworkResult struct {
	Message string                 `json:"message"`
	Result  *CmdbNetworkPageResult `json:"result"`
}

// query : cmdb 服务树节点串
// source ： 资源类型
// 参考：https://meicai.feishu.cn/docs/doccniuHjuYFzFQAT0kBO52e1kf?sidebarOpen=1#EDzlgz
func EndpointUnderNodeGets(url string, timeout int, query string, source string) ([]*dataobj.Endpoint, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	params := make(map[string]interface{})
	page := Pagination{
		PageNo:    1,
		PageSize:  100,
		TotalPage: 999,
	}
	params["sourceType"] = source
	params["expression"] = query

	ret := make([]*dataobj.Endpoint, 0)
	for page.PageNo <= page.TotalPage {
		params["pagination"] = page

		data, err := RequestByPost(url, timeout, params)
		if err != nil {
			logger.Errorf("request url %s params %s failed, %s", url, params, err)
			return nil, err
		}

		var pageRet Pagination
		switch source {
		case config.CmdbSourceHost:
			var hostResult CmdbHostResult
			err = json.Unmarshal(data, &hostResult)
			if err != nil || hostResult.Result == nil {
				logger.Errorf("Error: Parse %s JSON %v.", source, err)
				return nil, err
			}
			pageRet = hostResult.Result.Pagination
			ret = append(ret, convertHost2Endpoint(hostResult.Result.Hosts)...)
		case config.CmdbSourceNet:
			var networkResult CmdbNetworkResult
			err = json.Unmarshal(data, &networkResult)
			if err != nil || networkResult.Result == nil {
				logger.Errorf("Error: Parse %s JSON %v.", source, err)
				return nil, err
			}
			pageRet = networkResult.Result.Pagination
			ret = append(ret, convertNetwork2Endpoint(networkResult.Result.Networks)...)
		case config.CmdbSourceInst:
			var instResult CmdbInstanceResult
			err = json.Unmarshal(data, &instResult)
			if err != nil || instResult.Result == nil {
				logger.Errorf("Error: Parse %s JSON %v.", source, err)
				return nil, err
			}
			pageRet = instResult.Result.Pagination
			ret = append(ret, convertInstance2Endpoint(instResult.Result.Instances)...)
		}

		if pageRet.PageNo == 0 {
			return ret, fmt.Errorf("page result is nil")
		}
		page.PageNo++
		page.TotalPage = pageRet.TotalPage
	}

	return ret, nil
}

func convertHost2Endpoint(hosts []*CmdbHost) []*dataobj.Endpoint {
	if len(hosts) == 0 {
		return make([]*dataobj.Endpoint, 0)
	}

	ret := make([]*dataobj.Endpoint, 0)
	for _, host := range hosts {
		endpoint := &dataobj.Endpoint{
			Id:    host.Id,
			Ident: host.Ip,
			Alias: host.HostName,
		}
		extra := make(map[string]string, 0)
		extra["env"] = host.EnvCode
		extra["idc"] = host.DataCenterCode
		extra["type"] = host.Type
		endpoint.Tags = str.SortedTags(extra)

		ret = append(ret, endpoint)
	}
	return ret
}

func convertNetwork2Endpoint(networks []*Network) []*dataobj.Endpoint {
	if len(networks) == 0 {
		return make([]*dataobj.Endpoint, 0)
	}

	ret := make([]*dataobj.Endpoint, 0)
	for _, network := range networks {
		endpoint := &dataobj.Endpoint{
			Id:    network.Id,
			Ident: network.ManageIp,
			Alias: network.Name,
		}
		extra := make(map[string]string, 0)
		extra["env"] = network.EnvCode
		extra["idc"] = network.DataCenterCode
		extra["type"] = network.Type
		endpoint.Tags = str.SortedTags(extra)

		ret = append(ret, endpoint)
	}
	return ret
}

func convertInstance2Endpoint(instances []*Instance) []*dataobj.Endpoint {
	if len(instances) == 0 {
		return make([]*dataobj.Endpoint, 0)
	}

	ret := make([]*dataobj.Endpoint, 0)
	for _, instance := range instances {
		endpoint := &dataobj.Endpoint{
			Id:    instance.Id,
			Ident: instance.IP,
			Alias: instance.HostName,
		}
		extra := make(map[string]string, 0)
		extra["env"] = instance.EnvCode
		extra["group"] = instance.GroupCode
		extra["idc"] = instance.DataCenterCode
		extra["port"] = strconv.Itoa(instance.Port)
		extra["app"] = instance.AppCode

		endpoint.Tags = str.SortedTags(extra)

		ret = append(ret, endpoint)
	}
	return ret
}

func RequestByPost(url string, timeout int, params map[string]interface{}) ([]byte, error) {
	b, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	c := &http.Client{
		Timeout: time.Duration(timeout) * time.Millisecond,
	}

	resp, err := c.Do(req)
	if err != nil {
		logger.Errorf("Request post error %v.", err)
		return nil, err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("Request post Read Resp %v.", err)
		return nil, err
	}

	return data, err
}

func RequestByGet(url string, timeout int) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	c := &http.Client{
		Timeout: time.Duration(timeout) * time.Millisecond,
	}

	resp, err := c.Do(req)
	if err != nil {
		logger.Errorf("Request Get %v", err)
		return nil, err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("Request Get Read Resp %v", err)
		return nil, err
	}

	return data, err
}
