package dataobj

type App struct {
	Id        int64  `json:"id"`
	Code      string `json:"code"`
	Language  string `json:"language"`
	Name      string `json:"name"`
	SrvTreeId int64  `json:"srvTreeId"`
	Basic     bool   `json:"basic"`
}

type CmdbHost struct {
	Id             int64  `json:"id"`
	Ip             string `json:"innerIp"`
	HostName       string `json:"hostName"`
	Type           string `json:"type"`
	EnvCode        string `json:"envCode"`
	DataCenterCode string `json:"dataCenterCode"`
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

type Network struct {
	Id             int64  `json:"id"`
	ManageIp       string `json:"manageIp"`
	Name           string `json:"name"`
	DataCenterCode string `json:"dataCenterCode"`
	EnvCode        string `json:"envCode"`
	Type           string `json:"type"`
	SerialNo       string `json:"serialNo"`
}
