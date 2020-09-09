package routes

import (
	"github.com/didi/nightingale/src/dataobj"
	"github.com/didi/nightingale/src/model"
	"github.com/didi/nightingale/src/modules/monapi/cmdb"
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/errors"
)

func maskconfGetsHawkeye(c *gin.Context) {
	nid := urlParamInt64(c, "id")

	objs, err := model.MaskconfGetsHawkeye(nid)
	errors.Dangerous(err)

	for i := 0; i < len(objs); i++ {
		errors.Dangerous(objs[i].FillEndpoints())
	}

	renderData(c, objs, nil)
}

type tagKQueryForm struct {
	Nid    int64  `json:"nid"`
	Metric string `json:"metric"`
}

/**
leaf节点查询节点下endpoint的influxdb标签
其他节点查询influxdb表中全量标签
*/
func maskconfTagKeysPost(c *gin.Context) {
	var f tagKQueryForm
	errors.Dangerous(c.ShouldBind(&f))

	curNode, err := cmdb.GetCmdb().NodeGet("id", f.Nid)
	errors.Dangerous(err)

	res := []string{}
	qEndpoints := make([]string, 0)
	if curNode.Leaf == 1 {
		leafIds, err := cmdb.GetCmdb().LeafIds(curNode)
		errors.Dangerous(err)

		endpoints, err := cmdb.GetCmdb().EndpointUnderLeafs(leafIds)
		errors.Dangerous(err)

		for _, e := range endpoints {
			qEndpoints = append(qEndpoints, e.Ident)
		}
	}

	tagKResps, err := getTagKeysByTransfer(f.Metric, qEndpoints)
	errors.Dangerous(err)

	for _, resp := range tagKResps {
		if resp.Metric == f.Metric {
			for _, k := range resp.TagKeys {
				if k == "endpoint" {
					continue
				}
				res = append(res, k)
			}
		}
	}

	renderData(c, res, err)
}

type tagVQueryForm struct {
	Nid      int64              `json:"nid"`
	Metric   string             `json:"metric"`
	Include  []*dataobj.TagPair `json:"include"`
	Exclude  []*dataobj.TagPair `json:"exclude"`
	QueryKey string             `json:"queryKey"`
	QueryVal string             `json:"queryVal"`
	Limit    int                `json:"limit"`
}

/**
leaf节点查询节点下endpoint的influxdb标签
其他节点查询influxdb表中全量标签
*/
func maskconfTagValsPost(c *gin.Context) {
	var f tagVQueryForm
	errors.Dangerous(c.ShouldBind(&f))

	limit := 50
	if f.Limit > 0 {
		limit = f.Limit
	}

	if f.QueryKey == "" {
		errors.Dangerous("查询标签为空")
	}

	res := make([]string, 0)

	curNode, err := cmdb.GetCmdb().NodeGet("id", f.Nid)
	errors.Dangerous(err)

	leafNids, err := cmdb.GetCmdb().LeafIds(curNode)
	errors.Dangerous(err)

	qEndpoints := make([]string, 0)
	if curNode.Leaf == 1 {
		endpoints, _, err := cmdb.GetCmdb().EndpointUnderNodeGets(leafNids, "", "", "", 100, offset(c, f.Limit,
			999))
		errors.Dangerous(err)

		for _, e := range endpoints {
			qEndpoints = append(qEndpoints, e.Ident)
		}
	}

	req := dataobj.TagValsCludeRecv{
		Endpoints: qEndpoints,
		Metric:    f.Metric,
		Include:   f.Include,
		Exclude:   f.Exclude,
		QueryPair: []*dataobj.TagPair{{Key: f.QueryKey, Values: []string{f.QueryVal}}},
		Limit:     limit,
	}
	tagVResp, err := getTagValsByTransfer(req)
	errors.Dangerous(err)

	for _, tag := range tagVResp.Tagkvs {
		res = append(res, tag.Values...)
	}

	renderData(c, res, err)
}
