package config

const (
	Version        = 1
	RECOVERY       = "recovery"
	ALERT          = "alert"
	JudgesReplicas = 500
)

const (
	//根据服务树id获取服务树详情
	OpsSrvtreePath = "/srv_tree/"
	//根据服务树id获取子孙节点
	OpsSrvtreeDescendantsPath = "/srv_tree/descendants"
)

const (
	//获取用户详情信息
	SsoSearchUserPath = "/adminuser/searchadmin"
)
