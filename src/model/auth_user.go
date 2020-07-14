package model

type Result struct {
	Ret      int        `json:"ret"`
	AuthUser []AuthUser `json:"data"`
}
type AuthUser struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
	Phone  string `json:"phone"`
}
