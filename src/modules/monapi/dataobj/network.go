package dataobj

type Network struct {
	ID             int    `json:"id"`
	ManageIp       string `json:"manageIp"`
	Name           string `json:"name"`
	DataCenterCode string `json:"dataCenterCode"`
	EnvCode        string `json:"envCode"`
	Type           string `json:"type"`
	SerialNo       bool   `json:"serialNo"`
}

type NetworkPageResult struct {
	Pagination Pagination `json:"pagination"`
	Result     []*Network `json:"result"`
}

type NetworkResult struct {
	Message string             `json:"message"`
	Result  *NetworkPageResult `json:"result"`
}
