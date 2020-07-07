package backend

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/influxdata/influxdb/models"

	"github.com/didi/nightingale/src/dataobj"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/toolkits/pkg/logger"
)

// select value from metric where ...
func (influxdb *InfluxdbStorage) QueryData(inputs []dataobj.QueryData) []*dataobj.TsdbQueryResponse {
	logger.Debugf("query data, inputs: %+v", inputs)

	c, err := NewInfluxdbClient(influxdb.section)
	defer c.Client.Close()

	if err != nil {
		logger.Errorf("init influxdb client fail: %v", err)
		return nil
	}

	queryResponse := make([]*dataobj.TsdbQueryResponse, 0)
	for _, input := range inputs {
		for _, counter := range input.Counters {
			items := strings.Split(counter, "/")
			metric := items[0]
			tags := strings.Split(items[1], ",")
			influxdbQuery := InfluxdbQuery{
				Start:     input.Start,
				End:       input.End,
				Metric:    metric,
				Endpoints: input.Endpoints,
				Tags:      tags,
				Step:      input.Step,
				DsType:    input.DsType,
			}
			logger.Debugf("query influxql %s", influxdbQuery.RawQuery)

			query := client.NewQuery(influxdbQuery.RawQuery, c.Database, c.Precision)
			if response, err := c.Client.Query(query); err == nil && response.Error() == nil {
				for _, result := range response.Results {
					for _, series := range result.Series {

						// fixme : influx client get series.Tags is nil
						endpoint := series.Tags["endpoint"]
						delete(series.Tags, endpoint)
						counter, err := GetCounter(series.Name, "", series.Tags)
						if err != nil {
							logger.Warningf("get counter error: %+v", err)
							continue
						}
						values := convertValues(series)

						resp := &dataobj.TsdbQueryResponse{
							Start:    influxdbQuery.Start,
							End:      influxdbQuery.End,
							Endpoint: endpoint,
							Counter:  counter,
							DsType:   influxdbQuery.DsType,
							Step:     influxdbQuery.Step,
							Values:   values,
						}
						queryResponse = append(queryResponse, resp)
					}
				}
			}
		}
	}
	return queryResponse
}

// todo : 支持 comparison
// select value from metric where ...
func (influxdb *InfluxdbStorage) QueryDataForUI(input dataobj.QueryDataForUI) []*dataobj.TsdbQueryResponse {

	logger.Debugf("query data for ui, input: %+v", input)

	c, err := NewInfluxdbClient(influxdb.section)
	defer c.Client.Close()

	if err != nil {
		logger.Errorf("init influxdb client fail: %v", err)
		return nil
	}

	influxdbQuery := InfluxdbQuery{
		Start:     input.Start,
		End:       input.End,
		Metric:    input.Metric,
		Endpoints: input.Endpoints,
		Tags:      input.Tags,
		Step:      input.Step,
		DsType:    input.DsType,
		GroupKey:  input.GroupKey,
		AggrFunc:  input.AggrFunc,
	}
	influxdbQuery.renderSelect()
	influxdbQuery.renderEndpoints()
	influxdbQuery.renderTags()
	influxdbQuery.renderTimeRange()
	influxdbQuery.renderGroupBy()
	logger.Debugf("query influxql %s", influxdbQuery.RawQuery)

	queryResponse := make([]*dataobj.TsdbQueryResponse, 0)
	query := client.NewQuery(influxdbQuery.RawQuery, c.Database, c.Precision)
	if response, err := c.Client.Query(query); err == nil && response.Error() == nil {

		for _, result := range response.Results {
			for _, series := range result.Series {

				// fixme : influx client get series.Tags is nil
				endpoint := series.Tags["endpoint"]
				delete(series.Tags, endpoint)
				counter, err := GetCounter(series.Name, "", series.Tags)
				if err != nil {
					logger.Warningf("get counter error: %+v", err)
					continue
				}
				values := convertValues(series)

				resp := &dataobj.TsdbQueryResponse{
					Start:    influxdbQuery.Start,
					End:      influxdbQuery.End,
					Endpoint: endpoint,
					Counter:  counter,
					DsType:   influxdbQuery.DsType,
					Step:     influxdbQuery.Step,
					Values:   values,
				}
				queryResponse = append(queryResponse, resp)
			}
		}
	}
	return queryResponse
}

// show measurements on n9e
func (influxdb *InfluxdbStorage) QueryMetrics(recv dataobj.EndpointsRecv) *dataobj.MetricResp {
	logger.Debugf("query metric, recv: %+v", recv)

	c, err := NewInfluxdbClient(influxdb.section)
	defer c.Client.Close()

	if err != nil {
		logger.Errorf("init influxdb client fail: %v", err)
		return nil
	}

	influxql := fmt.Sprintf("SHOW MEASUREMENTS ON \"%s\"", influxdb.section.Database)
	query := client.NewQuery(influxql, c.Database, c.Precision)
	if response, err := c.Client.Query(query); err == nil && response.Error() == nil {
		resp := &dataobj.MetricResp{
			Metrics: make([]string, 0),
		}
		for _, result := range response.Results {
			for _, series := range result.Series {
				for _, valuePair := range series.Values {
					metric := valuePair[0].(string)
					resp.Metrics = append(resp.Metrics, metric)
				}
			}
		}
		return resp
	}
	return nil
}

// show tag keys / values from metric ...
func (influxdb *InfluxdbStorage) QueryTagPairs(recv dataobj.EndpointMetricRecv) []dataobj.IndexTagkvResp {
	logger.Debugf("query tag pairs, recv: %+v", recv)

	c, err := NewInfluxdbClient(influxdb.section)
	defer c.Client.Close()

	if err != nil {
		logger.Errorf("init influxdb client fail: %v", err)
		return nil
	}

	resp := make([]dataobj.IndexTagkvResp, 0)
	for _, metric := range recv.Metrics {
		tagkvResp := dataobj.IndexTagkvResp{
			Endpoints: recv.Endpoints,
			Metric:    metric,
			Tagkv:     make([]*dataobj.TagPair, 0),
		}
		// show tag keys
		influxql := fmt.Sprintf("SHOW TAG KEYS ON \"%s\" FROM \"%s\"", influxdb.section.Database, metric)
		query := client.NewQuery(influxql, c.Database, c.Precision)
		if response, err := c.Client.Query(query); err == nil && response.Error() == nil {
			keys := make([]string, 0)
			for _, result := range response.Results {
				for _, series := range result.Series {
					for _, valuePair := range series.Values {
						tagKey := valuePair[0].(string)
						// 去掉默认tag endpoint
						if tagKey != "endpoint" {
							keys = append(keys, tagKey)
						}
					}
				}
			}
			if len(keys) > 0 {
				// show tag values
				influxql := fmt.Sprintf("SHOW TAG VALUES ON \"%s\" FROM \"%s\" WITH KEY in (\"%s\")",
					influxdb.section.Database,
					metric, strings.Join(keys, "\",\""))
				query := client.NewQuery(influxql, c.Database, c.Precision)
				if response, err := c.Client.Query(query); err == nil && response.Error() == nil {
					tagPairs := make(map[string]*dataobj.TagPair)
					for _, result := range response.Results {
						for _, series := range result.Series {
							for _, valuePair := range series.Values {
								tagKey := valuePair[0].(string)
								tagValue := valuePair[1].(string)
								if pair, exist := tagPairs[tagKey]; exist {
									pair.Values = append(pair.Values, tagValue)
								} else {
									pair := &dataobj.TagPair{
										Key:    tagKey,
										Values: []string{tagValue},
									}
									tagPairs[pair.Key] = pair
									tagkvResp.Tagkv = append(tagkvResp.Tagkv, pair)
								}
							}
						}
					}
				}
			}
		}
		resp = append(resp, tagkvResp)
	}

	return resp
}

// show series from metric where ...
func (influxdb *InfluxdbStorage) QueryIndexByClude(recvs []dataobj.CludeRecv) []dataobj.XcludeResp {
	logger.Debugf("query IndexByClude , recv: %+v", recvs)

	c, err := NewInfluxdbClient(influxdb.section)
	defer c.Client.Close()

	if err != nil {
		logger.Errorf("init influxdb client fail: %v", err)
		return nil
	}
	resp := make([]dataobj.XcludeResp, 0)
	for _, recv := range recvs {
		xcludeResp := dataobj.XcludeResp{
			Endpoints: recv.Endpoints,
			Metric:    recv.Metric,
			Tags:      make([]string, 0),
			Step:      -1, // fixme
			DsType:    "GAUGE",
		}

		if len(recv.Endpoints) == 0 {
			resp = append(resp, xcludeResp)
			continue
		}

		// render endpoints
		endpointPart := "("
		for _, endpoint := range recv.Endpoints {
			endpointPart += fmt.Sprintf(" \"endpoint\"='%s' OR", endpoint)
		}
		endpointPart = endpointPart[:len(endpointPart)-len("OR")]
		endpointPart += ")"
		influxql := fmt.Sprintf("SHOW SERIES ON \"%s\" FROM \"%s\" WHERE %s ", influxdb.section.Database,
			recv.Metric, endpointPart)

		if len(recv.Include) > 0 {
			// include
			includePart := "("
			for _, include := range recv.Include {
				for _, value := range include.Values {
					includePart += fmt.Sprintf(" \"%s\"='%s' OR", include.Key, value)
				}
			}
			includePart = includePart[:len(includePart)-len("OR")]
			includePart += ")"
			influxql = fmt.Sprintf(" %s AND %s", influxql, includePart)
		}

		if len(recv.Exclude) > 0 {
			// exclude
			excludePart := "("
			for _, exclude := range recv.Exclude {
				for _, value := range exclude.Values {
					excludePart += fmt.Sprintf(" \"%s\"='%s' OR", exclude.Key, value)
				}
			}
			excludePart = excludePart[:len(excludePart)-len("OR")]
			excludePart += ")"
			influxql = fmt.Sprintf(" %s AND %s", influxql, excludePart)
		}

		query := client.NewQuery(influxql, c.Database, c.Precision)
		if response, err := c.Client.Query(query); err == nil && response.Error() == nil {
			for _, result := range response.Results {
				for _, series := range result.Series {
					for _, valuePair := range series.Values {

						// proc.port.listen,endpoint=localhost,port=22,service=sshd
						tagKey := valuePair[0].(string)

						// process
						items := strings.Split(tagKey, ",")
						newItems := make([]string, 0)
						for _, item := range items {
							if item != recv.Metric && !strings.Contains(item, "endpoint") {
								newItems = append(newItems, item)
							}
						}

						if len(newItems) > 0 {
							if tags, err := dataobj.SplitTagsString(strings.Join(newItems, ",")); err == nil {
								xcludeResp.Tags = append(xcludeResp.Tags, dataobj.SortedTags(tags))
							}
						}
					}
				}
			}
		}
		resp = append(resp, xcludeResp)
	}

	return resp
}

// show series from metric where ...
func (influxdb *InfluxdbStorage) QueryIndexByFullTags(recvs []dataobj.IndexByFullTagsRecv) []dataobj.
	IndexByFullTagsResp {
	logger.Debugf("query IndexByFullTags , recv: %+v", recvs)

	c, err := NewInfluxdbClient(influxdb.section)
	defer c.Client.Close()

	if err != nil {
		logger.Errorf("init influxdb client fail: %v", err)
		return nil
	}

	resp := make([]dataobj.IndexByFullTagsResp, 0)
	for _, recv := range recvs {
		fullTagResp := dataobj.IndexByFullTagsResp{
			Endpoints: recv.Endpoints,
			Metric:    recv.Metric,
			Tags:      make([]string, 0),
			Step:      -1, // FIXME
			DsType:    "GAUGE",
		}

		if len(recv.Endpoints) == 0 {
			resp = append(resp, fullTagResp)
			continue
		}

		// render endpoints
		endpointPart := ""
		for _, endpoint := range recv.Endpoints {
			endpointPart += fmt.Sprintf(" \"endpoint\"='%s' OR", endpoint)
		}
		endpointPart = endpointPart[:len(endpointPart)-len("OR")]
		influxql := fmt.Sprintf("SHOW SERIES ON \"%s\" FROM \"%s\" WHERE %s ", influxdb.section.Database,
			recv.Metric, endpointPart)

		query := client.NewQuery(influxql, c.Database, c.Precision)
		if response, err := c.Client.Query(query); err == nil && response.Error() == nil {
			for _, result := range response.Results {
				for _, series := range result.Series {
					for _, valuePair := range series.Values {

						// proc.port.listen,endpoint=localhost,port=22,service=sshd
						tagKey := valuePair[0].(string)

						// process
						items := strings.Split(tagKey, ",")
						newItems := make([]string, 0)
						for _, item := range items {
							if item != recv.Metric && !strings.Contains(item, "endpoint") {
								newItems = append(newItems, item)
							}
						}

						if len(newItems) > 0 {
							if tags, err := dataobj.SplitTagsString(strings.Join(newItems, ",")); err == nil {
								fullTagResp.Tags = append(fullTagResp.Tags, dataobj.SortedTags(tags))
							}
						}
					}
				}
			}
		}
		resp = append(resp, fullTagResp)
	}

	return resp
}

type InfluxdbQuery struct {
	Start     int64
	End       int64
	Metric    string
	Endpoints []string
	Tags      []string
	Step      int
	DsType    string
	GroupKey  []string //聚合维度
	AggrFunc  string   //聚合计算

	RawQuery string
}

func (query *InfluxdbQuery) renderSelect() {
	// select
	if query.AggrFunc != "" && len(query.GroupKey) > 0 {
		query.RawQuery = ""
	} else {
		query.RawQuery = fmt.Sprintf("SELECT \"value\" FROM \"%s\"", query.Metric)
	}
}

func (query *InfluxdbQuery) renderEndpoints() {
	// where endpoint
	if len(query.Endpoints) > 0 {
		endpointPart := "("
		for _, endpoint := range query.Endpoints {
			endpointPart += fmt.Sprintf(" \"endpoint\"='%s' OR", endpoint)
		}
		endpointPart = endpointPart[:len(endpointPart)-len("OR")]
		endpointPart += ")"
		query.RawQuery = fmt.Sprintf("%s WHERE %s", query.RawQuery, endpointPart)
	}
}

func (query *InfluxdbQuery) renderTags() {
	// where tags
	if len(query.Tags) > 0 {
		s := strings.Join(query.Tags, ",")
		tags, err := dataobj.SplitTagsString(s)
		if err != nil {
			logger.Warningf("split tags error, %+v", err)
			return
		}

		tagPart := "("
		for tagK, tagV := range tags {
			tagPart += fmt.Sprintf(" \"%s\"='%s' AND", tagK, tagV)
		}
		tagPart = tagPart[:len(tagPart)-len("AND")]
		tagPart += ")"

		if strings.Contains(query.RawQuery, "WHERE") {
			query.RawQuery = fmt.Sprintf("%s AND %s", query.RawQuery, tagPart)
		} else {
			query.RawQuery = fmt.Sprintf("%s WHERE %s", query.RawQuery, tagPart)
		}
	}
}

func (query *InfluxdbQuery) renderTimeRange() {
	// time
	if strings.Contains(query.RawQuery, "WHERE") {
		query.RawQuery = fmt.Sprintf("%s AND time >= %d AND time <= %d", query.RawQuery,
			time.Duration(query.Start)*time.Second,
			time.Duration(query.End)*time.Second)
	} else {
		query.RawQuery = fmt.Sprintf("%s WHERE time >= %d AND time <= %d", query.RawQuery, query.Start, query.End)
	}
}

func (query *InfluxdbQuery) renderGroupBy() {
	// group by
	if len(query.GroupKey) > 0 {
		groupByPart := strings.Join(query.GroupKey, ",")
		query.RawQuery = fmt.Sprintf("%s GROUP BY %s", query.RawQuery, groupByPart)
	}
}

func convertValues(series models.Row) []*dataobj.RRDData {

	// convert values
	values := make([]*dataobj.RRDData, 0, len(series.Values))
	for _, valuePair := range series.Values {
		timestampNumber, _ := valuePair[0].(json.Number)
		timestamp, _ := timestampNumber.Int64()

		valueNumber, _ := valuePair[1].(json.Number)
		valueFloat, _ := valueNumber.Float64()
		values = append(values, dataobj.NewRRDData(timestamp, valueFloat))
	}
	return values
}
