package stra

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/toolkits/pkg/logger"
	"github.com/toolkits/pkg/net/httplib"

	"github.com/didi/nightingale/src/model"
	"github.com/didi/nightingale/src/modules/collector/cache"
	"github.com/didi/nightingale/src/toolkits/address"
)

func GetSieves() {
	if !StraConfig.Enable {
		return
	}

	syncSieves()
	go loopSyncSieves()
}

func loopSyncSieves() {
	t1 := time.NewTicker(time.Duration(StraConfig.Interval) * time.Second)
	for {
		<-t1.C
		syncSieves()
	}
}

func syncSieves() {
	c, err := GetSievesRetry()
	if err != nil {
		logger.Errorf("get collect err:%v", err)
		return
	}

	cache.SieveCache.SetAll(c)
}

type SievesResp struct {
	Dat []model.ConfigInfo `json:"dat"`
	Err string             `json:"err"`
}

func GetSievesRetry() ([]model.ConfigInfo, error) {
	count := len(address.GetHTTPAddresses("monapi"))
	var resp SievesResp
	var err error
	for i := 0; i < count; i++ {
		resp, err = getSieves()
		if err == nil {
			if resp.Err != "" {
				err = fmt.Errorf(resp.Err)
				continue
			}
			return resp.Dat, err
		}
	}

	return resp.Dat, err
}

func getSieves() (SievesResp, error) {
	addrs := address.GetHTTPAddresses("monapi")
	i := rand.Intn(len(addrs))
	addr := addrs[i]

	var res SievesResp
	var err error

	url := fmt.Sprintf("http://%s%s", addr, StraConfig.SieveApi)
	err = httplib.Get(url).SetTimeout(time.Duration(StraConfig.Timeout) * time.Millisecond).ToJSON(&res)
	if err != nil {
		err = fmt.Errorf("get sieve config from remote:%s failed, error:%v", url, err)
	}

	return res, err
}
