package meicai

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/toolkits/pkg/logger"
)

type Pagination struct {
	PageNo      int `json:"pageNo"`
	PageSize    int `json:"pageSize"`
	Start       int `json:"start"`
	TotalPage   int `json:"totalPage"`
	TotalRecord int `json:"totalRecord"`
}

func RequestByPost(url string, timeout int, params map[string]interface{}) ([]byte, error) {
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
		Timeout: time.Duration(timeout) * time.Millisecond,
	}

	resp, err := c.Do(req)
	if err != nil {
		logger.Errorf("Request post error %v.", err)
		return nil, err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("Request post Read Resp %v.", err)
		return nil, err
	}

	return data, err
}

func RequestByGet(url string, timeout int) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	c := &http.Client{
		Timeout: time.Duration(timeout) * time.Millisecond,
	}

	resp, err := c.Do(req)
	if err != nil {
		logger.Errorf("Request Get %v", err)
		return nil, err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("Request Get Read Resp %v", err)
		return nil, err
	}

	return data, err
}
