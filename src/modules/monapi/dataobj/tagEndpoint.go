package dataobj

import (
	"fmt"
	"strings"
)

const (
	EndpointKeyPrefix  = "endpoint"
	EndpointKeyForLock = "hawkeye_endpoint_lock"
)

type TagEndpoint struct {
	Ip       string `json:"ip"`
	HostName string `json:"hostname"`
	EnvCode  string `json:"envCode"`
	Endpoint string `json:"endpoint"`
}

func BuildKey(srvType, srvTag string) string {
	return fmt.Sprintf("%s_%s_%s", EndpointKeyPrefix, srvType, srvTag)
}

func SplitKey(key string) (string, string) {
	if arr := strings.Split(key, "_"); len(arr) != 0 {
		return arr[1], arr[2] // srvType, srvTag
	}
	return "", ""
}
