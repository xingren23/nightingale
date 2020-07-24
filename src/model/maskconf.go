package model

import (
	"fmt"
	"strings"
)

type Maskconf struct {
	Id        int64    `json:"id"`
	Nid       int64    `json:"nid"`
	NodePath  string   `json:"node_path" xorm:"-"`
	Metric    string   `json:"metric"`
	Tags      string   `json:"tags"`
	Cause     string   `json:"cause"`
	User      string   `json:"user"`
	Btime     int64    `json:"btime"`
	Etime     int64    `json:"etime"`
	Endpoints []string `json:"endpoints" xorm:"-"`
}

func (mc *Maskconf) Add(endpoints []string) error {
	_, err := DB["mon"].Insert(mc)
	if err != nil {
		return err
	}

	affected := 0

	for i := 0; i < len(endpoints); i++ {
		endpoint := strings.TrimSpace(endpoints[i])
		if endpoint == "" {
			continue
		}

		_, err = DB["mon"].Insert(&MaskconfEndpoints{
			MaskId:   mc.Id,
			Endpoint: endpoint,
		})

		if err != nil {
			return err
		}

		affected++
	}

	if affected == 0 {
		return fmt.Errorf("arg[endpoints] empty")
	}

	return nil
}

func (mc *Maskconf) FillEndpoints() error {
	var objs []MaskconfEndpoints
	err := DB["mon"].Where("mask_id=?", mc.Id).OrderBy("id").Find(&objs)
	if err != nil {
		return err
	}

	cnt := len(objs)
	arr := make([]string, cnt)
	for i := 0; i < cnt; i++ {
		arr[i] = objs[i].Endpoint
	}

	mc.Endpoints = arr
	return nil
}

func MaskconfGets(nodeId int64, path string) ([]Maskconf, error) {
	var objs []Maskconf
	err := DB["mon"].Where("nid=?", nodeId).OrderBy("id desc").Find(&objs)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(objs); i++ {
		objs[i].NodePath = path
	}

	return objs, nil
}

func MaskconfDel(id int64) error {
	_, err := DB["mon"].Where("mask_id=?", id).Delete(new(MaskconfEndpoints))
	if err != nil {
		return err
	}

	_, err = DB["mon"].Where("id=?", id).Delete(new(Maskconf))
	return err
}

func MaskconfGetAll() ([]Maskconf, error) {
	var objs []Maskconf
	err := DB["mon"].Find(&objs)
	return objs, err
}

func CleanExpireMask(now int64) error {
	var objs []Maskconf
	err := DB["mon"].Where("etime<?", now).Cols("id").Find(&objs)
	if err != nil {
		return err
	}

	if len(objs) == 0 {
		return nil
	}

	session := DB["mon"].NewSession()
	defer session.Close()

	if err = session.Begin(); err != nil {
		return err
	}

	for i := 0; i < len(objs); i++ {
		if _, err := session.Exec("delete from maskconf where id=?", objs[i].Id); err != nil {
			session.Rollback()
			return err
		}

		if _, err := session.Exec("delete from maskconf_endpoints where mask_id=?", objs[i].Id); err != nil {
			session.Rollback()
			return err
		}
	}

	err = session.Commit()
	return err
}

func MaskconfGet(col string, value interface{}) (*Maskconf, error) {
	var obj Maskconf
	has, err := DB["mon"].Where(col+"=?", value).Get(&obj)
	if err != nil {
		return nil, err
	}

	if !has {
		return nil, nil
	}

	return &obj, nil
}

func (mc *Maskconf) Update(endpoints []string, cols ...string) error {
	session := DB["mon"].NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		return err
	}

	if _, err := session.Where("id=?", mc.Id).Cols(cols...).Update(mc); err != nil {
		session.Rollback()
		return err
	}

	if _, err := session.Exec("delete from maskconf_endpoints where mask_id=?", mc.Id); err != nil {
		session.Rollback()
		return err
	}

	affected := 0
	for i := 0; i < len(endpoints); i++ {
		endpoint := strings.TrimSpace(endpoints[i])
		if endpoint == "" {
			continue
		}

		_, err := session.Insert(&MaskconfEndpoints{
			MaskId:   mc.Id,
			Endpoint: endpoint,
		})

		if err != nil {
			session.Rollback()
			return err
		}

		affected += 1
	}

	if affected == 0 {
		session.Rollback()
		return fmt.Errorf("arg[endpoints] empty")
	}

	if err := session.Commit(); err != nil {
		session.Rollback()
		return err
	}

	return nil
}
