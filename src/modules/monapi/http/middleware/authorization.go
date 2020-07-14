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
			user := GetUser(c)
			username = user.Username
		}

		if username == "" {
			errors.Bomb("unauthorized")
		}

		c.Set("username", username)
		c.Next()
	}
}

func getCookie(c *gin.Context) string {
	//cookie, err := c.Request.Cookie("beta_user_token")
	//if err != nil {
	//	return ""
	//}

	//val, _ := url.QueryUnescape(cookie.Value)
	val, _ := url.QueryUnescape("%7B%22data%22%3A%7B%22id%22%3A%22201487%22%2C%22name%22%3A%22%E9%AB%98%E6%B3%A2%22%2C%22email%22%3A%22gaobo05%40meicai.cn%22%2C%22phone%22%3A%2213720059830%22%7D%7D")
	return val
}

func GetUser(c *gin.Context) *model.User {
	userStr := getCookie(c)
	if userStr == "" {
		errors.Bomb("login first please")
	}

	var userCookie User
	if err := json.Unmarshal([]byte(userStr), &userCookie); err != nil {
		errors.Bomb("login first please")
	}

	user, _ := model.UserGet("username", strings.Split(userCookie.Data.Email, "@")[0])
	if user == nil {
		user = &model.User{
			Username: userCookie.Data.Email,
			Dispname: userCookie.Data.Name,
			Phone:    userCookie.Data.Phone,
			Email:    userCookie.Data.Email,
			IsRoot:   1,
		}

		user.Save()
	}

	return user
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

type User struct {
	Data Data `json:"data"`
}

type Data struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}
