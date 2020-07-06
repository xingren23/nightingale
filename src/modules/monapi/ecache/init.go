package ecache

func Init() {
	EndpointCache = NewEndpointCache()
	SrvTreeCache = NewSrvTreeCache()
	SrvTagEndpointCache = NewSrvTagEndpointCache()
	MonitorItemCache = NewMonitorItemCache()
}
