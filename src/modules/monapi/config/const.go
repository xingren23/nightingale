package config

const (
	Version        = 1
	RECOVERY       = "recovery"
	ALERT          = "alert"
	JudgesReplicas = 500
)

const (
	//根据服务树id获取服务树详情
	OPS_GET_SRVTREE = "/srv_tree/"
	//根据服务树id获取子孙节点
	OPS_SRVTREE_DESCENDANTS = "/srv_tree/descendants"
)

const (
	//获取用户详情信息
	SSO_SEARCH_USER = "/adminuser/searchadmin"
)
