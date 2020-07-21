package ecache

var (
	AppCache      *AppCacheList
	HostCache     *HostCacheList
	InstanceCache *InstanceCacheList
	NetworkCache  *NetworkCacheList

	SrvTreeCache  *SrvTreeCacheMap
	EndpointCache *EndpointCacheMap
)

func Init() {
	// 服务树缓存
	SrvTreeCache = NewSrvTreeCache()
	EndpointCache = NewEndpointCache()
	// cmdb资源缓存
	HostCache = NewHostCache()
	AppCache = NewAppCache()
	InstanceCache = NewInstanceCache()
	NetworkCache = NewNetworkCache()
}
