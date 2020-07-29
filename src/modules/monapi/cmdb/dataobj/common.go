package dataobj

type Endpoint struct {
	Id    int64  `json:"id"`
	Ident string `json:"ident"`
	Alias string `json:"alias"`
	Tags  string `json:"tags"`
}

type Node struct {
	Id   int64  `json:"id"`
	Pid  int64  `json:"pid"`
	Name string `json:"name"`
	Path string `json:"path"`
	Leaf int    `json:"leaf"`
	Note string `json:"note"`
}

type EndpointBinding struct {
	Ident string `json:"ident"`
	Alias string `json:"alias"`
	Nodes []Node `json:"nodes"`
}
