package meicai

import (
	"encoding/json"
	"errors"

	"github.com/didi/nightingale/src/modules/monapi/config"
	"github.com/toolkits/pkg/logger"
)

type App struct {
	Id        int64  `json:"id"`
	Code      string `json:"code"`
	Language  string `json:"language"`
	Name      string `json:"name"`
	SrvTreeId int64  `json:"srvTreeId"`
	Basic     bool   `json:"basic"`
}

type AppPageResult struct {
	Pagination Pagination `json:"pagination"`
	Result     []*App     `json:"result"`
}

type AppResult struct {
	Message string         `json:"message"`
	Result  *AppPageResult `json:"result"`
}

func getAppByPage(url string, pageNo, pageSize int) (*AppPageResult, error) {
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
func GetAllApps() ([]*App, error) {
	res := []*App{}
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

func getHostByPage(url string, pageNo, pageSize int) (*CmdbHostPageResult, error) {
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
func GetAllHosts() ([]*CmdbHost, error) {
	res := []*CmdbHost{}
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

type Instance struct {
	Id             int64  `json:"id"`
	AppCode        string `json:"appCode"`
	AppId          int    `json:"appId"`
	DataCenterCode string `json:"dataCenterCode"`
	EnvCode        string `json:"envCode"`
	GroupCode      string `json:"groupCode"`
	IP             string `json:"ip"`
	HostName       string `json:"hostName"`
	Port           int    `json:"port"`
	UUID           string `json:"uuid"`
}

type InstPageResult struct {
	Pagination Pagination  `json:"pagination"`
	Result     []*Instance `json:"result"`
}

type InstResult struct {
	Message string          `json:"message"`
	Result  *InstPageResult `json:"result"`
}

func getInstanceByPage(url string, pageNo, pageSize int) (*InstPageResult, error) {
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
func GetAllInstances() ([]*Instance, error) {
	res := []*Instance{}
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

type Network struct {
	Id             int64  `json:"id"`
	ManageIp       string `json:"manageIp"`
	Name           string `json:"name"`
	DataCenterCode string `json:"dataCenterCode"`
	EnvCode        string `json:"envCode"`
	Type           string `json:"type"`
	SerialNo       string `json:"serialNo"`
}

type NetworkSearchResult struct {
	Message    string     `json:"message"`
	Pagination Pagination `json:"pagination"`
	Result     []*Network `json:"result"`
}

func getNetworkByPage(url string, pageNo, pageSize int) (*NetworkSearchResult, error) {
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
func GetAllNetworks() ([]*Network, error) {
	res := []*Network{}
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
