package routes

import (
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/errors"

	"github.com/didi/nightingale/src/model"
	"github.com/didi/nightingale/src/modules/monapi/cmdb"
)

type MaskconfForm struct {
	Nid       int64    `json:"nid"`
	Endpoints []string `json:"endpoints"`
	Metric    string   `json:"metric"`
	Tags      string   `json:"tags"`
	Cause     string   `json:"cause"`
	Btime     int64    `json:"btime"`
	Etime     int64    `json:"etime"`
}

func (f MaskconfForm) Validate() {
	mustNode(f.Nid)

	if f.Endpoints == nil || len(f.Endpoints) == 0 {
		errors.Bomb("arg[endpoints] empty")
	}

	if f.Btime >= f.Etime {
		errors.Bomb("args[btime,etime] invalid")
	}
}

func maskconfPost(c *gin.Context) {
	var f MaskconfForm
	errors.Dangerous(c.ShouldBind(&f))
	f.Validate()

	obj := &model.Maskconf{
		Nid:    f.Nid,
		Metric: f.Metric,
		Tags:   f.Tags,
		Cause:  f.Cause,
		Btime:  f.Btime,
		Etime:  f.Etime,
		User:   loginUsername(c),
	}

	renderMessage(c, obj.Add(f.Endpoints))
}

func maskconfGets(c *gin.Context) {
	nid := urlParamInt64(c, "id")

	node, err := cmdb.GetCmdb().NodeGet("id", nid)
	errors.Dangerous(err)

	maskconfs := make([]model.Maskconf, 0)
	if node.Leaf == 1 {
		// 查询当前节点
		objs, err := model.MaskconfGets(node.Id, node.Path)
		errors.Dangerous(err)
		maskconfs = append(maskconfs, objs...)
	} else {
		// 查询所有子节点
		nIds, err := cmdb.GetCmdb().LeafIds(node)
		errors.Dangerous(err)

		for _, nid := range nIds {
			node, err := cmdb.GetCmdb().NodeGet("id", nid)
			errors.Dangerous(err)

			objs, err := model.MaskconfGets(node.Id, node.Path)
			errors.Dangerous(err)
			maskconfs = append(maskconfs, objs...)
		}
	}

	for i := 0; i < len(maskconfs); i++ {
		errors.Dangerous(maskconfs[i].FillEndpoints())
	}

	sort.Slice(maskconfs, func(i int, j int) bool {
		if maskconfs[i].NodePath < maskconfs[j].NodePath {
			return true
		}
		if maskconfs[i].Id > maskconfs[j].Id {
			return true
		}
		return false
	})

	renderData(c, maskconfs, nil)
}

func maskconfDel(c *gin.Context) {
	id := urlParamInt64(c, "id")
	renderMessage(c, model.MaskconfDel(id))
}

func maskconfPut(c *gin.Context) {
	mc, err := model.MaskconfGet("id", urlParamInt64(c, "id"))
	errors.Dangerous(err)

	if mc == nil {
		errors.Bomb("maskconf is nil")
	}

	var f MaskconfForm
	errors.Dangerous(c.ShouldBind(&f))
	f.Nid = mc.Nid
	f.Validate()

	mc.Metric = f.Metric
	mc.Tags = f.Tags
	mc.Etime = f.Etime
	mc.Btime = f.Btime
	mc.Cause = f.Cause
	renderMessage(c, mc.Update(f.Endpoints, "metric", "tags", "etime", "btime", "cause"))
}
