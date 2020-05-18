package backend

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/didi/nightingale/src/dataobj"
	"github.com/didi/nightingale/src/modules/index/cache"
	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/toolkits/logger"
)

type InfluxClient struct {
	Client    client.Client
	Database  string
	Precision string
}

func NewInfluxClient(addr string) (*InfluxClient, error) {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     addr,
		Username: Config.Influxdb.Username,
		Password: Config.Influxdb.Password,
		Timeout:  time.Millisecond * time.Duration(Config.Influxdb.Timeout),
	})

	if err != nil {
		return nil, err
	}

	return &InfluxClient{
		Client:    c,
		Database:  Config.Influxdb.Database,
		Precision: Config.Influxdb.Precision,
	}, nil
}

func (c *InfluxClient) Send(items []*dataobj.InfluxDBItem) error {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  c.Database,
		Precision: c.Precision,
	})
	if err != nil {
		logger.Error("create batch points error: ", err)
		return err
	}

	for _, item := range items {
		pt, err := client.NewPoint(item.Measurement, item.Tags, item.Fields, time.Unix(item.Timestamp, 0))
		if err != nil {
			logger.Error("create new points error: ", err)
			continue
		}
		bp.AddPoint(pt)
	}

	return c.Client.Write(bp)
}

func (c *InfluxClient) QueryMetrics(endpoints []string) ([]string, error) {
	influxql := "show measurements where "
	for _, endpoint := range endpoints {
		influxql += fmt.Sprintf(" endpoint = '%s' or", endpoint)
	}
	metrics := make([]string, 0)
	influxql = influxql[:len(influxql)-2]
	q := client.NewQuery(influxql, Config.Influxdb.Database, Config.Influxdb.Precision)
	response, err := c.Client.Query(q)
	if err == nil && response.Error() == nil {
		for _, row := range response.Results[0].Series {
			for _, value := range row.Values {
				for _, item := range value {
					metrics = append(metrics, fmt.Sprintf("%s", item))
				}
			}
		}
		return metrics, nil
	} else {
		return nil, err
	}
}

func (c *InfluxClient) QueryMetricIndex(endpoints []string, measurement string) ([]*cache.TagPair, error) {
	tagPairs := make([]*cache.TagPair, 0)
	// show tag keys
	influxqlKeys := fmt.Sprintf("show tag keys from \"%s\" where ", measurement)
	for _, endpoint := range endpoints {
		influxqlKeys += fmt.Sprintf(" endpoint = '%s' or", endpoint)
	}
	influxqlKeys = influxqlKeys[:len(influxqlKeys)-2]

	tagKeys := make([]string, 0)
	queryKey := client.NewQuery(influxqlKeys, Config.Influxdb.Database, Config.Influxdb.Precision)
	if response, err := c.Client.Query(queryKey); err == nil && response.Error() == nil {
		for _, row := range response.Results[0].Series {
			for _, value := range row.Values {
				for _, item := range value {
					// 排除endpoint tag
					keyStr := fmt.Sprintf("%s", item)
					if keyStr != "endpoint" {
						tagKeys = append(tagKeys, keyStr)
					}
				}
			}
		}
	}
	if len(tagKeys) == 0 {
		return tagPairs, nil
	}

	// show tag values
	tagKeysStr := strings.Join(tagKeys, "\",\"")
	influxqlValues := fmt.Sprintf("show tag values from \"%s\" with key in (\"%s\") where ", measurement, tagKeysStr)
	for _, endpoint := range endpoints {
		influxqlValues += fmt.Sprintf(" endpoint = '%s' or", endpoint)
	}
	influxqlValues = influxqlValues[:len(influxqlValues)-2]

	tagkvs := make(map[string][]string)
	queryValue := client.NewQuery(influxqlValues, Config.Influxdb.Database, Config.Influxdb.Precision)
	if response, err := c.Client.Query(queryValue); err == nil && response.Error() == nil {
		for _, row := range response.Results[0].Series {
			for _, value := range row.Values {
				key := fmt.Sprintf("%s", value[0])
				if _, exist := tagkvs[key]; !exist {
					tagvs := make([]string, 0)
					tagkvs[key] = tagvs
				}
				tagkvs[key] = append(tagkvs[key], fmt.Sprintf("%s", value[1]))
			}
		}
	}

	for tagk, tagvs := range tagkvs {
		tagkv := &cache.TagPair{
			Key:    tagk,
			Values: tagvs,
		}
		tagPairs = append(tagPairs, tagkv)
	}
	return tagPairs, nil
}

func (c *InfluxClient) QueryIndexByFullTags(endpoints []string, measurement string, tagkvs []*cache.TagPair) (tags []string, err error) {
	// show tag keys
	influxql := fmt.Sprintf("show series from \"%s\" where ", measurement)
	tagKvStr := make([]string, 0)
	for _, tagkv := range tagkvs {
		var keyStr string
		key := tagkv.Key
		for _, value := range tagkv.Values {
			keyStr += fmt.Sprintf(" %s = '%s' or", key, value)
		}
		if len(keyStr) > 2 {
			keyStr = fmt.Sprintf("( %s )", keyStr[:len(keyStr)-2])
		}
		if len(keyStr) > 2 {
			tagKvStr = append(tagKvStr, keyStr)
		}
	}

	var endpointStr string
	for _, endpoint := range endpoints {
		endpointStr += fmt.Sprintf(" endpoint = '%s' or", endpoint)
	}
	if len(endpointStr) > 2 {
		endpointStr = fmt.Sprintf("( %s )", endpointStr[:len(endpointStr)-2])
	}

	if len(endpointStr) > 2 {
		tagKvStr = append(tagKvStr, endpointStr)
	}
	influxql += fmt.Sprintf("%s", strings.Join(tagKvStr, " and "))

	queryKey := client.NewQuery(influxql, Config.Influxdb.Database, Config.Influxdb.Precision)
	tagsMap := make(map[string]string, 0)
	var response *client.Response
	response, err = c.Client.Query(queryKey)
	if err == nil && response.Error() == nil {
		for _, row := range response.Results[0].Series {
			for _, value := range row.Values {
				for _, item := range value {
					// 去掉measurement 和 endpoint tag
					keyStr := fmt.Sprintf("%s", item)
					tagkvItems := strings.Split(keyStr, ",")[1:]

					tagsStr := make([]string, 0)
					for _, tagkvItem := range tagkvItems {
						if !strings.HasPrefix(tagkvItem, "endpoint") {
							tagsStr = append(tagsStr, tagkvItem)
						}
					}
					sort.Strings(tagsStr)
					if len(tagsStr) > 0 {
						item := strings.Join(tagsStr, ",")
						if _, exist := tagsMap[item]; !exist {
							tags = append(tags, item)
							tagsMap[item] = item
						}
					}
				}
			}
		}
	}
	return
}

func (c *InfluxClient) QueryData(start, end int64, endpoint, metric, dstype string, tags map[string]string) (*dataobj.TsdbQueryResponse, error) {
	timeStr := fmt.Sprintf("( time >= %d and time <= %d )", start*time.Second.Nanoseconds(),
		end*time.Second.Nanoseconds())
	endpointStr := fmt.Sprintf(" ( endpoint = '%s' ) ", endpoint)
	influxql := fmt.Sprintf("select %s from \"%s\" where %s and %s",
		strings.ToLower(dstype), metric, endpointStr, timeStr)

	// TODO: tags
	influxql += fmt.Sprintf(" group by *")

	q := client.NewQuery(influxql, Config.Influxdb.Database, Config.Influxdb.Precision)
	response, err := c.Client.Query(q)
	if err == nil && response.Error() == nil {
		for _, result := range response.Results {
			for _, row := range result.Series {
				endpoint := row.Tags["endpoint"]
				delete(row.Tags, "endpoint")
				tagStr := dataobj.SortedTags(row.Tags)
				counter := dataobj.PKWithTags(row.Name, tagStr)
				tsdbResponse := &dataobj.TsdbQueryResponse{
					Start:    start,
					End:      end,
					DsType:   dstype,
					Counter:  counter,
					Endpoint: endpoint,
					Values:   make([]*dataobj.RRDData, 0),
				}
				for _, value := range row.Values {
					ts, errTs := value[0].(json.Number).Int64()
					val, errVal := value[1].(json.Number).Float64()
					if errTs == nil && errVal == nil {
						tsdbResponse.Values = append(tsdbResponse.Values, dataobj.NewRRDData(ts, val))
					}
				}
				return tsdbResponse, nil
			}
		}
	}
	return nil, err
}
