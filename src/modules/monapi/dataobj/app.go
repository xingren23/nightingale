package dataobj

type App struct {
	ID        int    `json:"id"`
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
