package meicai

import (
	"strings"
	"time"

	"github.com/didi/nightingale/src/model"

	"github.com/didi/nightingale/src/modules/monapi/config"
	"github.com/toolkits/pkg/errors"
	"github.com/toolkits/pkg/net/httplib"
)

type AuthResp struct {
	Ret      int        `json:"ret"`
	AuthUser []AuthUser `json:"data"`
}
type AuthUser struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
	Phone  string `json:"phone"`
}

func SaveSSOUser(userNames []string) ([]int64, error) {
	cnt := len(userNames)
	ret := make([]int64, 0, cnt)

	for _, userName := range userNames {
		user, err := model.UserGet("username", userName)
		if err != nil {
			return nil, err
		}

		if user == nil {
			url := config.Get().Api.SSO + config.SSO_SEARCH_USER

			m := map[string]string{
				"email": userName,
			}

			var resp AuthResp
			err := httplib.Post(url).JSONBodyQuiet(m).SetTimeout(3 * time.Second).ToJSON(&resp)
			if err != nil {
				return nil, err
			}

			if resp.AuthUser == nil {
				return nil, err
			}

			if len(resp.AuthUser) == 0 {
				errors.Bomb("用户[%v]不存在: %v", userName)
			}

			authUser := resp.AuthUser[0]
			if authUser.Status == "0" {
				errors.Bomb("用户[%v]为禁用状态: %v", userName)
			}

			user = &model.User{
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
