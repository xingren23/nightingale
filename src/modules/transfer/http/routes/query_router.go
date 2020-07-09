package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/errors"
	"github.com/toolkits/pkg/logger"

	"github.com/didi/nightingale/src/dataobj"
	"github.com/didi/nightingale/src/modules/transfer/backend"
	"github.com/didi/nightingale/src/toolkits/http/render"
	"github.com/didi/nightingale/src/toolkits/stats"
)

type QueryDataReq struct {
	queryData []dataobj.QueryData
}

func QueryData(c *gin.Context) {
	stats.Counter.Set("data.api.qp10s", 1)

	storage, err := backend.GetStorageFor("")
	if err != nil {
		logger.Warningf("Could not find storage ")
		render.Message(c, err)
		return
	}

	var queryDataReq QueryDataReq
	errors.Dangerous(c.ShouldBindJSON(&queryDataReq))
	resp := storage.QueryData(queryDataReq.queryData)
	render.Data(c, resp, nil)
}

func QueryDataForUI(c *gin.Context) {
	stats.Counter.Set("data.ui.qp10s", 1)
	var input dataobj.QueryDataForUI
	var respData []*dataobj.QueryDataForUIResp
	errors.Dangerous(c.ShouldBindJSON(&input))
	start := input.Start
	end := input.End

	storage, err := backend.GetStorageFor("")
	if err != nil {
		logger.Warningf("Could not find storage ")
		render.Message(c, err)
		return
	}
	resp := storage.QueryDataForUI(input)
	for _, d := range resp {
		data := &dataobj.QueryDataForUIResp{
			Start:    d.Start,
			End:      d.End,
			Endpoint: d.Endpoint,
			Counter:  d.Counter,
			DsType:   d.DsType,
			Step:     d.Step,
			Values:   d.Values,
		}
		respData = append(respData, data)
	}

	if len(input.Comparisons) > 1 {
		for i := 1; i < len(input.Comparisons); i++ {
			comparison := input.Comparisons[i]
			input.Start = start - comparison
			input.End = end - comparison
			res := storage.QueryDataForUI(input)
			for _, d := range res {
				for j := range d.Values {
					d.Values[j].Timestamp += comparison
				}

				data := &dataobj.QueryDataForUIResp{
					Start:      d.Start,
					End:        d.End,
					Endpoint:   d.Endpoint,
					Counter:    d.Counter,
					DsType:     d.DsType,
					Step:       d.Step,
					Values:     d.Values,
					Comparison: comparison,
				}
				respData = append(respData, data)
			}
		}
	}

	render.Data(c, respData, nil)
}

func GetMetrics(c *gin.Context) {
	stats.Counter.Set("metric.qp10s", 1)
	recv := dataobj.EndpointsRecv{}
	errors.Dangerous(c.ShouldBindJSON(&recv))

	storage, err := backend.GetStorageFor("")
	if err != nil {
		logger.Warningf("Could not find storage ")
		render.Message(c, err)
		return
	}

	resp := storage.QueryMetrics(recv)

	render.Data(c, resp, nil)
}

func GetTagPairs(c *gin.Context) {
	stats.Counter.Set("tag.qp10s", 1)
	recv := dataobj.EndpointMetricRecv{}
	errors.Dangerous(c.ShouldBindJSON(&recv))

	storage, err := backend.GetStorageFor("")
	if err != nil {
		logger.Warningf("Could not find storage ")
		render.Message(c, err)
		return
	}

	resp := storage.QueryTagPairs(recv)
	render.Data(c, resp, nil)
}

func GetIndexByClude(c *gin.Context) {
	stats.Counter.Set("xclude.qp10s", 1)
	recvs := make([]dataobj.CludeRecv, 0)
	errors.Dangerous(c.ShouldBindJSON(&recvs))

	storage, err := backend.GetStorageFor("")
	if err != nil {
		logger.Warningf("Could not find storage ")
		render.Message(c, err)
		return
	}

	resp := storage.QueryIndexByClude(recvs)
	render.Data(c, resp, nil)
}

func GetIndexByFullTags(c *gin.Context) {
	stats.Counter.Set("counter.qp10s", 1)
	recvs := make([]dataobj.IndexByFullTagsRecv, 0)
	errors.Dangerous(c.ShouldBindJSON(&recvs))

	storage, err := backend.GetStorageFor("")
	if err != nil {
		logger.Warningf("Could not find storage ")
		render.Message(c, err)
		return
	}

	resp := storage.QueryIndexByFullTags(recvs)
	render.Data(c, resp, nil)
}
