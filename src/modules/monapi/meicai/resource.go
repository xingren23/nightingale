package meicai

import (
	"errors"

	"github.com/didi/nightingale/src/dataobj"

	jsoniter "github.com/json-iterator/go"

	"github.com/didi/nightingale/src/modules/monapi/config"
	"github.com/toolkits/pkg/logger"
)

type AppPageResult struct {
	Pagination Pagination     `json:"pagination"`
	Result     []*dataobj.App `json:"result"`
}

type AppResult struct {
	Message string         `json:"message"`
	Result  *AppPageResult `json:"result"`
}

func getAppByPage(url string, pageNo, pageSize int) (*AppPageResult, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	params := make(map[string]interface{})

	params["pagination"] = map[string]int{
		"pageNo":   pageNo,
		"pageSize": pageSize,
	}

	data, err := RequestByPost(url, params)
	if err != nil {
		return nil, err
	}

	var m AppPageResult
	err = json.Unmarshal(data, &m)
	if err != nil {
		logger.Errorf("Error: Parse app JSON %v.", err)
		return nil, err
	}

	return &m, nil
}

// 获取应用
func GetAllApps() ([]*dataobj.App, error) {
	res := []*dataobj.App{}
	url := config.Get().Api.OpsAddr + config.OpsAppSearchPath
	pageNo := 1
	pageSize := 100
	pageTotal := 999
	for pageNo <= pageTotal {
		page, err := getAppByPage(url, pageNo, pageSize)
		if err != nil {
			return res, err
		}
		if page == nil {
			return res, errors.New("get app page result is nil")
		}
		pageNo++
		pageTotal = page.Pagination.TotalPage
		res = append(res, page.Result...)
	}

	return res, nil
}

type CmdbHostPageResult struct {
	Pagination Pagination          `json:"pagination"`
	Hosts      []*dataobj.CmdbHost `json:"result"`
}

type CmdbHostResult struct {
	Message string              `json:"message"`
	Result  *CmdbHostPageResult `json:"result"`
}

func getHostByPage(url string, pageNo, pageSize int) (*CmdbHostPageResult, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	params := make(map[string]interface{})
	params["pageNo"] = pageNo
	params["pageSize"] = pageSize

	data, err := RequestByPost(url, params)
	if err != nil {
		return nil, err
	}

	var m CmdbHostPageResult
	err = json.Unmarshal(data, &m)
	if err != nil {
		logger.Errorf("Error: Parse cmdbHost JSON %v.", err)
		return nil, err
	}

	return &m, nil
}

// 获取主机
func GetAllHosts() ([]*dataobj.CmdbHost, error) {
	res := []*dataobj.CmdbHost{}
	url := config.Get().Api.OpsAddr + config.OpsHostSearchPath
	pageNo := 1
	pageSize := 100
	pageTotal := 999
	for pageNo <= pageTotal {
		page, err := getHostByPage(url, pageNo, pageSize)
		if err != nil {
			return res, err
		}
		if page == nil {
			return res, errors.New("get host page result is nil")
		}
		pageNo++
		pageTotal = page.Pagination.TotalPage
		res = append(res, page.Hosts...)
	}

	return res, nil
}

type InstPageResult struct {
	Pagination Pagination          `json:"pagination"`
	Result     []*dataobj.Instance `json:"result"`
}

type InstResult struct {
	Message string          `json:"message"`
	Result  *InstPageResult `json:"result"`
}

func getInstanceByPage(url string, pageNo, pageSize int) (*InstPageResult, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	params := make(map[string]interface{})
	params["pageNo"] = pageNo
	params["pageSize"] = pageSize

	data, err := RequestByPost(url, params)
	if err != nil {
		return nil, err
	}

	var m InstPageResult
	err = json.Unmarshal(data, &m)
	if err != nil {
		logger.Errorf("Error:Parse Instance JSON %v.", err)
		return nil, err
	}

	return &m, nil
}

// 获取实例
func GetAllInstances() ([]*dataobj.Instance, error) {
	res := []*dataobj.Instance{}
	url := config.Get().Api.OpsAddr + config.OpsInstanceSearchPath
	pageNo := 1
	pageSize := 100
	pageTotal := 999
	for pageNo <= pageTotal {
		page, err := getInstanceByPage(url, pageNo, pageSize)
		if err != nil {
			return res, err
		}
		if page == nil {
			return res, errors.New("page instance result is nil")
		}
		pageNo++
		pageTotal = page.Pagination.TotalPage
		res = append(res, page.Result...)
	}

	return res, nil
}

type NetworkSearchResult struct {
	Message    string             `json:"message"`
	Pagination Pagination         `json:"pagination"`
	Result     []*dataobj.Network `json:"result"`
}

func getNetworkByPage(url string, pageNo, pageSize int) (*NetworkSearchResult, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	params := make(map[string]interface{})
	params["pageNo"] = pageNo
	params["pageSize"] = pageSize

	data, err := RequestByPost(url, params)
	if err != nil {
		return nil, err
	}

	var m NetworkSearchResult
	err = json.Unmarshal(data, &m)
	if err != nil {
		logger.Errorf("Error: Parse network JSON %v.", err)
		return nil, err
	}

	return &m, nil
}

// 获取网络设备
func GetAllNetworks() ([]*dataobj.Network, error) {
	res := []*dataobj.Network{}
	url := config.Get().Api.OpsAddr + config.OpsNetworkSearchPath
	pageNo := 1
	pageSize := 100
	pageTotal := 999
	for pageNo <= pageTotal {
		page, err := getNetworkByPage(url, pageNo, pageSize)
		if err != nil {
			return res, err
		}
		if page == nil {
			return res, errors.New("page result is nil")
		}
		pageNo++
		pageTotal = page.Pagination.TotalPage
		res = append(res, page.Result...)
	}

	return res, nil
}
