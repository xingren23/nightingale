package dataobj

// 策略过滤标签
const (
	FilterTagNodePath = "nodePath"
	FilterTagHost     = "host"
	FilterTagEnv      = "env"
)

// SrvTagEndpointCache 缓存key
const (
	EndpointKeyDocker  = "docker"
	EndpointKeyPM      = "pm"
	EndpointKeyNetwork = "network"
)

const (
	CmdbSourceInst = "instance"
	CmdbSourceApp  = "app"
	CmdbSourceNet  = "network"
	CmdbSourceHost = "host"
)
