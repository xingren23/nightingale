package model

import (
	"fmt"
)

func (t *Team) ModifyHawkeye(ident, name string, mgmt int, admins, members []int64) error {
	adminIds, err := safeUserIds(admins)
	if err != nil {
		return err
	}

	memberIds, err := safeUserIds(members)
	if err != nil {
		return err
	}

	if len(adminIds) == 0 && len(memberIds) == 0 {
		return fmt.Errorf("no invalid memeber ids")
	}

	if mgmt == 1 && len(adminIds) == 0 {
		return fmt.Errorf("arg[admins] is necessary")
	}

	// 如果ident有变化，就要检查是否有重名
	if ident != t.Ident {
		cnt, err := DB["uic"].Where("ident = ? and nid = ? and id <> ?", ident, t.Nid, t.Id).Count(new(Team))
		if err != nil {
			return err
		}

		if cnt > 0 {
			return fmt.Errorf("ident[%s] already exists", ident)
		}
	}

	t.Ident = ident
	t.Name = name
	t.Mgmt = mgmt

	if err = t.CheckFields(); err != nil {
		return err
	}

	session := DB["uic"].NewSession()
	defer session.Close()

	if err = session.Begin(); err != nil {
		return err
	}

	if _, err = session.Where("id=?", t.Id).Cols("ident", "name", "mgmt").Update(t); err != nil {
		session.Rollback()
		return err
	}

	if _, err = session.Exec("DELETE FROM team_user WHERE team_id=?", t.Id); err != nil {
		session.Rollback()
		return err
	}

	for i := 0; i < len(adminIds); i++ {
		if err = teamUserBind(session, t.Id, adminIds[i], 1); err != nil {
			session.Rollback()
			return err
		}
	}

	for i := 0; i < len(memberIds); i++ {
		if err = teamUserBind(session, t.Id, memberIds[i], 0); err != nil {
			session.Rollback()
			return err
		}
	}

	return session.Commit()
}

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
