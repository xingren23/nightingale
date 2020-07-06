package dataobj

type CmdbHost struct {
	ID             int    `json:"id"`
	Ip             string `json:"innerIp"`
	HostName       string `json:"hostName"`
	Type           string `json:"type"`
	EnvCode        string `json:"envCode"`
	DataCenterCode string `json:"dataCenterCode"`
}

type CmdbHostPageResult struct {
	Pagination Pagination  `json:"pagination"`
	Result     []*Instance `json:"result"`
}

type CmdbHostResult struct {
	Message string              `json:"message"`
	Result  *CmdbHostPageResult `json:"result"`
}
