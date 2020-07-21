package mcache

var (
	MaskCache *MaskCacheMap
	StraCache *StraCacheMap

	EndpointCache    *EndpointCacheMap
	MonitorItemCache *MonitorItemCacheMap
)

func Init() {
	MaskCache = NewMaskCache()
	StraCache = NewStraCache()

	EndpointCache = NewEndpointCache()
	// 元数据缓存
	MonitorItemCache = NewMonitorItemCache()
}
