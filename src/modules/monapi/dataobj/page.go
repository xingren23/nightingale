package dataobj

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type Pagination struct {
	PageNo      int `json:"pageNo"`
	PageSize    int `json:"pageSize"`
	Start       int `json:"start"`
	TotalPage   int `json:"totalPage"`
	TotalRecord int `json:"totalRecord"`
}

func GetByOps(url string, params map[string]interface{}) ([]byte, error) {
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
		// todo
		//Timeout: time.Duration(g.Config().OpsApi.Timeout) * time.Millisecond,
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
