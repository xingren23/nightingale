package config

const (
	Version        = 1
	RECOVERY       = "recovery"
	ALERT          = "alert"
	JudgesReplicas = 500
)

const (
	OpsSrvtreeRootPath        = "/srv_tree/tree"
	OpsSrvtreePath            = "/srv_tree"
	OpsSrvtreeDescendantsPath = "/srv_tree/descendants"
	OpsSrvtreeAncestorsPath   = "/srv_tree/ancestors"

	OpsApiResourcerPath   = "/api/resource/query"
	OpsAppSearchPath      = "/app/search"
	OpsHostSearchPath     = "/host/search"
	OpsInstanceSearchPath = "/instance/search"
	OpsNetworkSearchPath  = "/network_device/search"
)

const (
	//获取用户详情信息
	SsoSearchUserPath = "/adminuser/searchadmin"
)

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
