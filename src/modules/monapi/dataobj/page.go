package dataobj

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/didi/nightingale/src/modules/monapi/config"
)

const (
	SourceInst = "instance"
	SourceApp  = "app"
	SourceNet  = "network"
	SourceHost = "host"
)

type Pagination struct {
	PageNo      int `json:"pageNo"`
	PageSize    int `json:"pageSize"`
	Start       int `json:"start"`
	TotalPage   int `json:"totalPage"`
	TotalRecord int `json:"totalRecord"`
}

func RequestByPost(url string, params map[string]interface{}) ([]byte, error) {
	b, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	c := &http.Client{
		Timeout: time.Duration(config.Get().Api.Timeout) * time.Millisecond,
	}

	resp, err := c.Do(req)
	if err != nil {
		log.Printf("Error :Cache Instance Request %v.\n", err)
		return nil, err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error :Cache Instance Read Resp %v.\n", err)
		return nil, err
	}

	return data, err
}

func RequestByGet(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	c := &http.Client{
		Timeout: time.Duration(config.Get().Api.Timeout) * time.Millisecond,
	}

	resp, err := c.Do(req)
	if err != nil {
		log.Printf("Error :Cache Instance Request %v.\n", err)
		return nil, err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error :Cache Instance Read Resp %v.\n", err)
		return nil, err
	}

	return data, err
}
