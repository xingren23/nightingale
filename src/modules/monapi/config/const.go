package config

const (
	Version        = 1
	RECOVERY       = "recovery"
	ALERT          = "alert"
	JudgesReplicas = 500
)

// 策略过滤标签
const (
	FilterTagNodePath = "nodePath"
	FilterTagHost     = "host"
	FilterTagEnv      = "env"
)

// 服务树资源类型 endpoint type
const (
	EndpointKeyDocker  = "docker"
	EndpointKeyPM      = "pm"
	EndpointKeyNetwork = "network"
	// todo 兼容app_instance
	EndpointKeyInstance = "instance"
)
