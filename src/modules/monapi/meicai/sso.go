package meicai

import (
	"fmt"
	"strings"
	"time"

	"github.com/toolkits/pkg/logger"

	"github.com/didi/nightingale/src/model"

	"github.com/didi/nightingale/src/modules/monapi/config"
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
			url := config.Get().Api.SSOAddr + config.SsoSearchUserPath

			m := map[string]string{
				"email": userName,
			}

			var resp AuthResp
			err := httplib.Post(url).JSONBodyQuiet(m).SetTimeout(time.Duration(config.Get().Api.Timeout) * time.
				Millisecond).ToJSON(
				&resp)
			if err != nil {
				logger.Errorf("request sso %s error, %s", url, err)
				return nil, err
			}

			if resp.AuthUser == nil {
				return nil, fmt.Errorf("request sso resp authUser is nil, url, %s", url)
			}

			if len(resp.AuthUser) == 0 {
				return nil, fmt.Errorf("用户[%v]不存在", userName)
			}

			authUser := resp.AuthUser[0]
			if authUser.Status == "0" {
				return nil, fmt.Errorf("用户[%v]为禁用状态", userName)
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
