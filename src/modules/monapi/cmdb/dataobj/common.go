package dataobj

import "strings"

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

type AppInstance struct {
	Id     int64  `json:"id"`
	App    string `json:"app"`
	Env    string `json:"env"`
	Group  string `json:"group"`
	Ident  string `json:"ident"`
	Port   int    `json:"port"`
	Tags   string `json:"tags"`
	NodeId int64  `json:"nodeId"`
}

func Paths(longPath string) []string {
	names := strings.Split(longPath, ".")
	count := len(names)
	paths := make([]string, 0, count)

	for i := 1; i <= count; i++ {
		paths = append(paths, strings.Join(names[:i], "."))
	}

	return paths
}
