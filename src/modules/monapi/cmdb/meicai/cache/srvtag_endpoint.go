package cache

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/didi/nightingale/src/modules/monapi/cmdb/dataobj"

	jsoniter "github.com/json-iterator/go"

	"github.com/didi/nightingale/src/modules/monapi/redisc"
	"github.com/toolkits/pkg/logger"
)

const (
	EndpointKeyPrefix  = "endpoint"
	EndpointKeyDot     = "#"
	EndpointKeyForLock = "hawkeye_endpoint_lock"
)

var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func RedisSrvTagKey(srvType, srvTag string) string {
	ret := bufferPool.Get().(*bytes.Buffer)
	ret.Reset()
	defer bufferPool.Put(ret)

	ret.WriteString(EndpointKeyPrefix)
	ret.WriteString(EndpointKeyDot)
	ret.WriteString(srvType)
	ret.WriteString(EndpointKeyDot)
	ret.WriteString(srvTag)
	return ret.String()
}

func SplitRedisKey(key string) (string, string) {
	if arr := strings.Split(key, EndpointKeyDot); len(arr) != 0 {
		return arr[1], arr[2] // srvType, srvTag
	}
	return "", ""
}

// 服务树节点串 + 资源类型 -> endpoint列表
func ScanRedisEndpointKeys() ([]string, error) {
	ret := []string{}
	var cursor int
	batch := 20
	for {
		key := fmt.Sprintf("%s%s*", EndpointKeyPrefix, EndpointKeyDot)
		data, err := redisc.SCAN(cursor, key, batch)
		if err != nil {
			return ret, nil
		}
		cursor, err = strconv.Atoi(string(data[0].([]uint8)))
		if err != nil {
			return ret, nil
		}
		for _, d := range data[1].([]interface{}) {
			key := string(d.([]uint8))
			ret = append(ret, key)
		}
		if cursor == 0 {
			break
		}
	}

	return ret, nil
}

func GetEndpointsFromRedis(srvType, srvTag string) ([]*dataobj.Endpoint, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	ret := make([]*dataobj.Endpoint, 0)
	key := RedisSrvTagKey(srvType, srvTag)
	bs, err := redisc.SMEMBERS(key)
	if err != nil {
		logger.Warningf("smembers %s failed", key)
		return ret, err
	}
	if len(bs) == 0 {
		logger.Warningf("key %s not in redis cache.", key)
		return ret, nil
	}
	for _, b := range bs {
		var te dataobj.Endpoint
		err := json.Unmarshal(b, &te)
		if err != nil {
			return ret, err
		}
		ret = append(ret, &te)
	}
	return ret, nil
}

func SetEndpointForRedis(key string, endpoints []*dataobj.Endpoint) error {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	batch := 10
	size := len(endpoints)
	n := size/batch + 1
	if size%batch == 0 {
		n = size / batch
	}
	for i := 0; i < n; i++ {
		tmp := []interface{}{}
		start := i * batch
		end := start + batch
		if i == n-1 && size%batch != 0 {
			end = start + size%batch
		}
		for _, e := range endpoints[start:end] {
			if e == nil {
				logger.Errorf("tag endpoint is nil")
				continue
			}
			data, err := json.Marshal(e)
			if err != nil {
				return err
			}
			tmp = append(tmp, data)
		}
		// 写redis
		redisc.SADD(key, tmp)
	}
	return nil
}

func SetEndpointLock() (bool, error) {
	return redisc.SETNX(EndpointKeyForLock, "1", 60*60)
}

func SetEndpointUnLock() {
	if ret := redisc.GET(EndpointKeyForLock); ret != 1 {
		logger.Info("%s key is not exists")
		return
	}

	if err := redisc.DelKey(EndpointKeyForLock); err != nil {
		logger.Errorf("unlock EndpointKeyForLock key error %v", err)
	}
}
