package mcache

var (
	MaskCache *MaskCacheMap
	StraCache *StraCacheMap

	MetricInfoCache *MetricInfoCacheMap
)

func Init() {
	MaskCache = NewMaskCache()
	StraCache = NewStraCache()

	// 元数据缓存
	MetricInfoCache = NewMetricInfoCache()
}
