package mcache

var (
	MaskCache *MaskCacheMap
	StraCache *StraCacheMap

	MonitorItemCache *MonitorItemCacheMap
)

func Init() {
	MaskCache = NewMaskCache()
	StraCache = NewStraCache()

	// 元数据缓存
	MonitorItemCache = NewMonitorItemCache()
}
