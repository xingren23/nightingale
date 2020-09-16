package config

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/didi/nightingale/src/toolkits/identity"
	"github.com/spf13/viper"
	"github.com/toolkits/pkg/file"
)

type Config struct {
	Salt     string                   `yaml:"salt"`
	Logger   loggerSection            `yaml:"logger"`
	HTTP     httpSection              `yaml:"http"`
	LDAP     ldapSection              `yaml:"ldap"`
	Redis    redisSection             `yaml:"redis"`
	Queue    queueSection             `yaml:"queue"`
	Cleaner  cleanerSection           `yaml:"cleaner"`
	Link     linkSection              `yaml:"link"`
	Notify   map[string][]string      `yaml:"notify"`
	Tokens   []string                 `yaml:"tokens"`
	SSO      ssoSection               `yaml:"sso"`
	Cmdb     CmdbSection              `yaml:"cmdb"`
	Identity identity.IdentitySection `yaml:"identity"`
}

type CmdbSection struct {
	Default string        `yaml:"default"`
	N9e     N9eSection    `yaml:"n9e"`
	Meicai  MeicaiSection `yaml:"meicai"`
}

type N9eSection struct {
	Enabled bool   `yaml:"enabled"`
	Name    string `yaml:"name"`
}

type MeicaiSection struct {
	Enabled bool   `yaml:"enabled"`
	Name    string `yaml:"name"`
	Timeout int    `yaml:"timeout"`
	OpsAddr string `yaml:"opsAddr"`
}

type ssoSection struct {
	CookieName string `yaml:"cookieName"`
	Timeout    int    `yaml:"timeout"`
	SSOAddr    string `yaml:"ssoAddr"`
}

type linkSection struct {
	Stra  string `yaml:"stra"`
	Event string `yaml:"event"`
	Claim string `yaml:"claim"`
}

type queueSection struct {
	EventPrefix  string        `yaml:"eventPrefix"`
	EventQueues  []interface{} `yaml:"-"`
	Callback     string        `yaml:"callback"`
	SenderPrefix string        `yaml:"senderPrefix"`
}

type cleanerSection struct {
	Days  int `yaml:"days"`
	Batch int `yaml:"batch"`
}

type redisSection struct {
	Addr    string         `yaml:"addr"`
	Pass    string         `yaml:"pass"`
	DB      int            `yaml:"db"`
	Idle    int            `yaml:"idle"`
	Timeout timeoutSection `yaml:"timeout"`
}

type timeoutSection struct {
	Conn  int `yaml:"conn"`
	Read  int `yaml:"read"`
	Write int `yaml:"write"`
}

type loggerSection struct {
	Dir       string `yaml:"dir"`
	Level     string `yaml:"level"`
	KeepHours uint   `yaml:"keepHours"`
}

type httpSection struct {
	Secret string `yaml:"secret"`
}

type ldapSection struct {
	Host            string         `yaml:"host"`
	Port            int            `yaml:"port"`
	BaseDn          string         `yaml:"baseDn"`
	BindUser        string         `yaml:"bindUser"`
	BindPass        string         `yaml:"bindPass"`
	AuthFilter      string         `yaml:"authFilter"`
	Attributes      ldapAttributes `yaml:"attributes"`
	CoverAttributes bool           `yaml:"coverAttributes"`
	AutoRegist      bool           `yaml:"autoRegist"`
	TLS             bool           `yaml:"tls"`
	StartTLS        bool           `yaml:"startTLS"`
}

type ldapAttributes struct {
	Dispname string `yaml:"dispname"`
	Phone    string `yaml:"phone"`
	Email    string `yaml:"email"`
	Im       string `yaml:"im"`
}

var (
	yaml *Config
)

// Get configuration file
func Get() *Config {
	return yaml
}

// Parse configuration file
func Parse(ymlfile string) error {
	bs, err := file.ReadBytes(ymlfile)
	if err != nil {
		return fmt.Errorf("cannot read yml[%s]: %v", ymlfile, err)
	}

	viper.SetConfigType("yaml")
	err = viper.ReadConfig(bytes.NewBuffer(bs))
	if err != nil {
		return fmt.Errorf("cannot read yml[%s]: %v", ymlfile, err)
	}

	viper.SetDefault("redis.idle", 4)
	viper.SetDefault("redis.timeout", map[string]int{
		"conn":  500,
		"read":  3000,
		"write": 3000,
	})

	viper.SetDefault("queue", map[string]string{
		"eventPrefix":  "/n9e/event/",
		"callback":     "/n9e/event/callback",
		"senderPrefix": "/n9e/sender/",
	})

	viper.SetDefault("cmdb.default", "n9e")
	viper.SetDefault("cmdb.n9e.enabled", true)
	viper.SetDefault("cmdb.n9e.name", "n9e")
	viper.SetDefault("cmdb.meicai.enabled", false)
	viper.SetDefault("cmdb.meicai.name", "meicai")

	viper.SetDefault("cleaner", map[string]int{
		"days":  366,
		"batch": 100,
	})

	var c Config
	err = viper.Unmarshal(&c)
	if err != nil {
		return fmt.Errorf("cannot read yml[%s]: %v", ymlfile, err)
	}

	size := len(c.Notify)
	if size == 0 {
		return fmt.Errorf("config.notify invalid")
	}

	prios := make([]string, size)
	i := 0
	for elt := range c.Notify {
		prios[i] = elt
		i++
	}

	sort.Strings(prios)

	prefix := c.Queue.EventPrefix
	for i := 0; i < size; i++ {
		c.Queue.EventQueues = append(c.Queue.EventQueues, prefix+prios[i])
	}

	yaml = &c

	return nil
}
