package model

import "time"

type ConfigInfo struct {
	Id         int64     `json:"id"`
	CfgGroup   string    `json:"cfg_group"`
	CfgKey     string    `json:"cfg_key"`
	CfgValue   string    `json:"cfg_value"`
	CreateTime time.Time `json:"create_time"`
	CreateBy   int64     `json:"create_by"`
	UpdateTime time.Time `json:"update_time"`
	UpdateBy   int64     `json:"update_by"`
	Status     int       `json:"status"`
}

func (c *ConfigInfo) Add() error {
	_, err := DB["mon"].Insert(c)
	return err
}

func (c *ConfigInfo) Update(col ...string) error {
	_, err := DB["mon"].Where("id=?", c.Id).Cols(col...).Update(c)
	return err
}

func (c *ConfigInfo) Del() error {
	_, err := DB["mon"].Where("id=?", c.Id).Delete(c)
	return err
}

func ConfigInfoGets(query string, limit, offest int) ([]*ConfigInfo, error) {
	session := DB["mon"].Limit(limit, offest).Desc("id")
	if query != "" {
		q := "%" + query + "%"
		session = session.Where("cfg_group like ? or cfg_key like ? and status > -1", q, q)
	}

	var items []*ConfigInfo
	err := session.Find(&items)
	return items, err
}

func ConfigInfoGet(col string, val interface{}) (*ConfigInfo, error) {
	var obj ConfigInfo
	has, err := DB["mon"].Where(col+"=?", val).Get(&obj)
	if err != nil {
		return nil, err
	}

	if !has {
		return nil, nil
	}

	return &obj, nil
}

func ConfigInfoTotal(query string) (int64, error) {
	if query != "" {
		q := "%" + query + "%"
		return DB["mon"].Where("cfg_group like ? or cfg_key like ?", q, q).Count(new(ConfigInfo))
	}
	return DB["mon"].Count(new(ConfigInfo))
}

func ConfigInfoGetByQ(group, key string) ([]*ConfigInfo, error) {
	session := DB["mon"].Where("status > -1 ").OrderBy("id")
	if group != "" {
		session = session.Where("cfg_group = ?", group)
	}

	if key != "" {
		session = session.Where("cfg_key = ?", key)
	}

	var items []*ConfigInfo
	err := session.Find(&items)
	return items, err
}
