package middleware

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strings"

	"github.com/didi/nightingale/src/model"
	"github.com/didi/nightingale/src/modules/monapi/config"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/errors"
	"github.com/toolkits/pkg/slice"
)

func Logined() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := cookieUser(c)
		if username == "" {
			username = headerUser(c)
		}

		if username == "" {
			username = devOpsTokenUser(c)
		}

		if username == "" {
			errors.Bomb("unauthorized")
		}

		c.Set("username", username)
		c.Next()
	}
}

func devOpsTokenUser(c *gin.Context) string {
	// TODO: user token
	//cookie, err := c.Request.Cookie(config.Get().Cookie.Name)
	//if err != nil {
	//	return ""
	//}

	userStr, _ := url.QueryUnescape("%7B%22data%22%3A%7B%22id%22%3A%22201487%22%2C%22name%22%3A%22%E9%AB%98%E6%B3%A2%22%2C%22email%22%3A%22gaobo05%40meicai.cn%22%2C%22phone%22%3A%2213720059830%22%7D%7D")
	//userStr, _ := url.QueryUnescape(cookie.Value)
	if userStr == "" {
		errors.Bomb("login first please")
	}

	var opsUserResp DevOpsUserResp
	if err := json.Unmarshal([]byte(userStr), &opsUserResp); err != nil {
		errors.Bomb("login first please")
	}

	// 自动创建用户
	user, _ := model.UserGet("username", strings.Split(opsUserResp.Data.Email, "@")[0])
	if user == nil {
		// TODO : 都是 root 用户？
		user = &model.User{
			Username: strings.Split(opsUserResp.Data.Email, "@")[0],
			Dispname: opsUserResp.Data.Name,
			Phone:    opsUserResp.Data.Phone,
			Email:    opsUserResp.Data.Email,
			IsRoot:   1,
		}
		user.Save()
	}

	return user.Dispname
}

func cookieUser(c *gin.Context) string {
	session := sessions.Default(c)

	value := session.Get("username")
	if value == nil {
		return ""
	}

	return value.(string)
}

func headerUser(c *gin.Context) string {
	auth := c.GetHeader("Authorization")

	if auth == "" {
		return ""
	}

	arr := strings.Fields(auth)
	if len(arr) != 2 {
		return ""
	}

	identity, err := base64.StdEncoding.DecodeString(arr[1])
	if err != nil {
		return ""
	}

	pair := strings.Split(string(identity), ":")
	if len(pair) != 2 {
		return ""
	}

	err = model.PassLogin(pair[0], pair[1])
	if err != nil {
		return ""
	}

	return pair[0]
}

const internalToken = "monapi-builtin-token"

// CheckHeaderToken check thirdparty x-srv-token
func CheckHeaderToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("x-srv-token")
		if token != internalToken && !slice.ContainsString(config.Get().Tokens, token) {
			errors.Bomb("token[%s] invalid", token)
		}
		c.Next()
	}
}

type DevOpsUserResp struct {
	Data DevOpsUser `json:"data"`
}

type DevOpsUser struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}
