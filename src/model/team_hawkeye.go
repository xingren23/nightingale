package model

import (
	"fmt"
)

func TeamHawkeyeAdd(ident, name string, mgmt int, members []int64, nid int64) error {
	memberIds, err := safeUserIds(members)
	if err != nil {
		return err
	}

	if len(memberIds) == 0 {
		return fmt.Errorf("no invalid memeber ids")
	}

	t := Team{
		Ident: ident,
		Name:  name,
		Mgmt:  mgmt,
		Nid:   nid,
	}

	if err = t.CheckFields(); err != nil {
		return err
	}

	session := DB["uic"].NewSession()
	defer session.Close()

	cnt, err := session.Where("ident=? and nid=?", ident, nid).Count(new(Team))
	if err != nil {
		return err
	}

	if cnt > 0 {
		return fmt.Errorf("%s already exists", ident)
	}

	if err = session.Begin(); err != nil {
		return err
	}

	if _, err = session.Insert(&t); err != nil {
		session.Rollback()
		return err
	}

	for i := 0; i < len(memberIds); i++ {
		if err := teamUserBind(session, t.Id, memberIds[i], 0); err != nil {
			session.Rollback()
			return err
		}
	}

	return session.Commit()
}

func TeamHawkeyeTotal(nids []int64, ids []int64) (int64, error) {
	if ids != nil && len(ids) > 0 {
		return DB["uic"].In("nid", nids).In("id", ids).Count(new(Team))

	}

	return DB["uic"].In("nid", nids).Count(new(Team))
}

func TeamHawkeyeGets(nids []int64, ids []int64, limit, offset int) ([]*Team, error) {
	session := DB["uic"].Limit(limit, offset).OrderBy("ident")
	if ids != nil {
		session = session.In("nid", nids).In("id", ids)
	} else {
		session = session.In("nid", nids)
	}

	var objs []*Team
	err := session.Find(&objs)
	return objs, err
}
