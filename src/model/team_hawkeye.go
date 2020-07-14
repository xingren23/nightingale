package model

import (
	"fmt"
	"github.com/didi/nightingale/src/modules/monapi/config"
	"github.com/toolkits/pkg/errors"
	"github.com/toolkits/pkg/net/httplib"
	"strings"
	"time"
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

	cnt, err := session.Where("ident=?", ident).Count(new(Team))
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

func SaveSSOUser(userNames []string) ([]int64, error) {
	cnt := len(userNames)
	ret := make([]int64, 0, cnt)

	for _, userName := range userNames {
		user, err := UserGet("username", userName)
		if err != nil {
			return nil, err
		}

		if user == nil {
			url := config.Get().Api.SSO + config.SSO_SEARCH_USER

			m := map[string]string{
				"email": userName,
			}

			var result Result
			err := httplib.Post(url).JSONBodyQuiet(m).SetTimeout(3 * time.Second).ToJSON(&result)
			if err != nil {
				return nil, err
			}

			if result.AuthUser == nil {
				return nil, err
			}

			authUser := result.AuthUser[0]
			if authUser.Status == "0" {
				errors.Bomb("cannot retrieve user[%d]: %v", userName, err)
			}

			user = &User{
				Username: strings.Split(authUser.Email, "@")[0],
				Password: "",
				Dispname: authUser.Name,
				Phone:    authUser.Phone,
				Email:    authUser.Email,
				IsRoot:   1,
			}

			user.Save()
		}

		ret = append(ret, user.Id)
	}

	return ret, nil
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
