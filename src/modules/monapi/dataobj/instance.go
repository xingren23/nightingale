package dataobj

type Instance struct {
	AppCode        string `json:"appCode"`
	AppID          int    `json:"appId"`
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
