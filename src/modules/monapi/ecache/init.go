package ecache

func Init() {
	// 服务树缓存
	SrvTreeCache = NewSrvTreeCache()
	EndpointCache = NewEndpointCache()

	// 元数据缓存
	MonitorItemCache = NewMonitorItemCache()

	// cmdb资源缓存
	HostCache = NewHostCache()
	AppCache = NewAppCache()
	InstanceCache = NewInstanceCache()
	NetworkCache = NewNetworkCache()
}
