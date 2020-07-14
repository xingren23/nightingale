package dataobj

import (
	"bytes"
	"strings"
	"sync"
)

const (
	EndpointKeyPrefix  = "endpoint"
	EndpointKeyDot     = "#"
	EndpointKeyForLock = "hawkeye_endpoint_lock"
)

type TagEndpoint struct {
	Ip       string `json:"ip"`
	HostName string `json:"hostname"`
	EnvCode  string `json:"envCode"`
	Endpoint string `json:"endpoint"`
}

var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func BuildKey(srvType, srvTag string) string {
	ret := bufferPool.Get().(*bytes.Buffer)
	ret.Reset()
	defer bufferPool.Put(ret)

	ret.WriteString(EndpointKeyPrefix)
	ret.WriteString(EndpointKeyDot)
	ret.WriteString(srvType)
	ret.WriteString(EndpointKeyDot)
	ret.WriteString(srvTag)
	return ret.String()
}

func SplitKey(key string) (string, string) {
	if arr := strings.Split(key, EndpointKeyDot); len(arr) != 0 {
		return arr[1], arr[2] // srvType, srvTag
	}
	return "", ""
}
