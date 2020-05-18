package routes

import (
	"fmt"
	"net/http/httputil"
	"net/url"

	"github.com/didi/nightingale/src/modules/index/cache"
	"github.com/didi/nightingale/src/modules/transfer/backend"
	"github.com/didi/nightingale/src/toolkits/address"
	"github.com/didi/nightingale/src/toolkits/http/render"
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/errors"
	"github.com/toolkits/pkg/logger"
)

type EndpointsRecv struct {
	Endpoints []string `json:"endpoints"`
}

type MetricList struct {
	Metrics []string `json:"metrics"`
}

func ProxyGetMetrics(c *gin.Context) {
	if backend.Config.Tsdb.Enabled {
		indexReq(c)
	} else if backend.Config.Influxdb.Enabled {
		recv := EndpointsRecv{}
		errors.Dangerous(c.ShouldBindJSON(&recv))

		resp := MetricList{}
		var err error
		for _, item := range backend.Config.Influxdb.ClusterList {
			for _, addr := range item.Addrs {
				client, err := backend.NewInfluxClient(addr)
				if err != nil {
					logger.Warningf("get influx client %s error, %v", addr, err)
				}
				metrics, err := client.QueryMetrics(recv.Endpoints)
				resp.Metrics = append(resp.Metrics, metrics...)
			}
		}

		render.Data(c, resp, err)
	} else {
		render.Data(c, nil, fmt.Errorf("backend tsdb all disabled"))
	}
}

type EndpointMetricRecv struct {
	Endpoints []string `json:"endpoints"`
	Metrics   []string `json:"metrics"`
}

type IndexTagkvResp struct {
	Endpoints []string         `json:"endpoints"`
	Metric    string           `json:"metric"`
	Tagkv     []*cache.TagPair `json:"tagkv"`
}

func ProxyGetTagPairs(c *gin.Context) {
	if backend.Config.Tsdb.Enabled {
		indexReq(c)
	} else if backend.Config.Influxdb.Enabled {
		recv := EndpointMetricRecv{}
		errors.Dangerous(c.ShouldBindJSON(&recv))

		resp := make([]*IndexTagkvResp, 0)
		for _, metric := range recv.Metrics {

			node, _ := backend.InfluxNodeRing.GetNode(metric)
			addr := backend.Config.Influxdb.Cluster[node]
			client, err := backend.NewInfluxClient(addr)
			if err != nil {
				logger.Warningf("get influx client %s error, %v", addr, err)
				continue
			}

			tagkvs, err := client.QueryMetricIndex(recv.Endpoints, metric)
			if err != nil {
				logger.Warningf("query metric index error, %v %s %v", recv.Endpoints, metric, err)
				continue
			}

			TagkvResp := IndexTagkvResp{
				Endpoints: recv.Endpoints,
				Metric:    metric,
				Tagkv:     tagkvs,
			}
			resp = append(resp, &TagkvResp)
		}

		render.Data(c, resp, nil)
	} else {
		render.Data(c, nil, fmt.Errorf("backend tsdb all disabled"))
	}
}

type GetIndexByFullTagsRecv struct {
	Endpoints []string         `json:"endpoints"`
	Metric    string           `json:"metric"`
	Tagkv     []*cache.TagPair `json:"tagkv"`
}

type GetIndexByFullTagsResp struct {
	Endpoints []string `json:"endpoints"`
	Metric    string   `json:"metric"`
	Tags      []string `json:"tags"`
	Step      int      `json:"step"`
	DsType    string   `json:"dstype"`
}

func ProxyGetIndexByFullTags(c *gin.Context) {
	if backend.Config.Tsdb.Enabled {
		indexReq(c)
	} else if backend.Config.Influxdb.Enabled {
		recv := make([]GetIndexByFullTagsRecv, 0)
		errors.Dangerous(c.ShouldBindJSON(&recv))

		var resp []GetIndexByFullTagsResp

		for _, r := range recv {
			endpoinsts := r.Endpoints
			metric := r.Metric
			tagkv := r.Tagkv

			node, _ := backend.InfluxNodeRing.GetNode(metric)
			addr := backend.Config.Influxdb.Cluster[node]
			client, err := backend.NewInfluxClient(addr)
			if err != nil {
				logger.Warningf("get influx client %s error, %v", addr, err)
			}

			tagsList, err := client.QueryIndexByFullTags(endpoinsts, metric, tagkv)

			resp = append(resp, GetIndexByFullTagsResp{
				Endpoints: r.Endpoints,
				Metric:    r.Metric,
				Tags:      tagsList,
				Step:      10,
				DsType:    "GAUGE",
			})
		}
		render.Data(c, resp, nil)
	} else {
		render.Data(c, nil, fmt.Errorf("backend tsdb all disabled"))
	}
}

type CludeRecv struct {
	Endpoints []string         `json:"endpoints"`
	Metric    string           `json:"metric"`
	Include   []*cache.TagPair `json:"include"`
	Exclude   []*cache.TagPair `json:"exclude"`
}

type XcludeResp struct {
	Endpoint string   `json:"endpoint"`
	Metric   string   `json:"metric"`
	Tags     []string `json:"tags"`
	Step     int      `json:"step"`
	DsType   string   `json:"dstype"`
}

func ProxyGetIndexByClude(c *gin.Context) {
	if backend.Config.Tsdb.Enabled {
		indexReq(c)
	} else if backend.Config.Influxdb.Enabled {
		//recv := make([]CludeRecv, 0)
		//errors.Dangerous(c.ShouldBindJSON(&recv))
		//
		//var resp []XcludeResp

	} else {
		render.Data(c, nil, fmt.Errorf("backend tsdb all disabled"))
	}
}

func indexReq(c *gin.Context) {
	target, err := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", address.GetHTTPPort("index")))
	errors.Dangerous(err)

	proxy := httputil.NewSingleHostReverseProxy(target)
	c.Request.Header.Set("X-Forwarded-Host", c.Request.Header.Get("Host"))

	proxy.ServeHTTP(c.Writer, c.Request)
}
