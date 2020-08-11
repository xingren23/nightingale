package meicai

import (
	"fmt"
	"log"
	"path"
	"strconv"
	"time"

	dataobj2 "github.com/didi/nightingale/src/dataobj"

	"github.com/didi/nightingale/src/modules/monapi/cmdb/dataobj"

	"github.com/toolkits/pkg/file"
	"github.com/toolkits/pkg/runner"
	"xorm.io/core"
	"xorm.io/xorm"

	"github.com/toolkits/pkg/logger"
)

type Meicai struct {
	DB      map[string]*xorm.Engine
	Timeout int
	OpsAddr string
}

const (
	OpsSrvtreeRootPath  = "/srv_tree/tree"
	OpsApiResourcerPath = "/api/resource/query"
)

type MySQLConf struct {
	Addr  string `yaml:"addr"`
	Max   int    `yaml:"max"`
	Idle  int    `yaml:"idle"`
	Debug bool   `yaml:"debug"`
}

func (meicai *Meicai) Init() {
	confdir := path.Join(runner.Cwd, "etc")

	mysqlYml := path.Join(confdir, "mysql.local.yml")
	if !file.IsExist(mysqlYml) {
		mysqlYml = path.Join(confdir, "mysql.yml")
	}

	confs := make(map[string]MySQLConf)
	err := file.ReadYaml(mysqlYml, &confs)
	if err != nil {
		log.Fatalf("cannot read yml[%s]: %v", mysqlYml, err)
	}
	name := "mon"
	conf, has := confs[name]
	if !has {
		log.Fatalf("no such mysql conf: %s", name)
	}

	db, err := xorm.NewEngine("mysql", conf.Addr)
	if err != nil {
		log.Fatalf("cannot connect mysql[%s]: %v", conf.Addr, err)
		panic(err)
	}

	db.SetMaxIdleConns(conf.Idle)
	db.SetMaxOpenConns(conf.Max)
	db.SetConnMaxLifetime(time.Hour)
	db.ShowSQL(conf.Debug)
	db.Logger().SetLevel(core.LOG_INFO)

	meicai.DB = map[string]*xorm.Engine{}
	meicai.DB[name] = db

	// 定时全量同步
	go meicai.SyncOpsLoop()
}

func (meicai *Meicai) SyncOpsLoop() {
	duration := time.Hour * time.Duration(24)
	for {
		// sync srvtree & endpoint
		err := meicai.SyncOps()
		if err != nil {
			logger.Errorf("sync meicai node failed, %s", err)
		}
		time.Sleep(duration)
	}
}

// init nodes & endpoints from ops
func (meicai *Meicai) SyncOps() error {
	start := time.Now()
	logger.Info("start init ops")
	nodes, err := meicai.InitNode()
	if err != nil {
		logger.Errorf("get nodes failed, %s", err)
		return err
	}
	// 遍历节点
	for _, node := range nodes {
		//初始化叶子节点的资源
		if node.Leaf == 1 {
			logger.Infof("init leaf node endpoint, id=%d path=%s", node.Id, node.Path)
			nodeStr := node.Path
			url := fmt.Sprintf("%s%s", meicai.OpsAddr, OpsApiResourcerPath)

			// 主机资源
			meicai.initNodeHosts(url, nodeStr, node.Id)
			// 网络资源
			meicai.initNodeNetworks(url, nodeStr, node.Id)

			// instance & app
			apps, err := QueryAppByNode(url, meicai.Timeout, nodeStr)
			if err != nil {
				appMap := make(map[string]*App, 0)
				for _, app := range apps {
					appMap[app.Code] = app
				}
				// 实例资源
				meicai.initNodeInstances(url, nodeStr, node.Id, appMap)
			}

			//time.Sleep(time.Duration(time.Millisecond) * 100)
		}
	}
	logger.Infof("end init ops, elapsed %s", time.Since(start))
	return nil
}

// InitNode 初始化服务树节点
func (m *Meicai) InitNode() ([]*dataobj.Node, error) {
	// get srvtree
	url := fmt.Sprintf("%s%s", m.OpsAddr, OpsSrvtreeRootPath)
	nodes, err := SrvTreeGets(url, m.Timeout)
	if err != nil {
		logger.Errorf("get srvtree failed, %s", err)
		return nil, err
	}

	err = m.commitNodes(nodes)
	logger.Info("srvtree node init done")
	return nodes, err
}

func (meicai *Meicai) commitNodes(nodes []*dataobj.Node) error {
	start := time.Now()
	session := meicai.DB["mon"].NewSession()
	defer session.Close()

	for _, node := range nodes {
		has, err := session.Exist(&dataobj.Node{Id: node.Id})
		if err != nil || !has {
			logger.Infof("insert node %v", node)
			_, err = session.Insert(node)
			if err != nil {
				logger.Errorf("insert node %v failed, %s", node, err)
				session.Rollback()
				return err
			}
		} else {
			logger.Infof("update node %v", node)
			_, err = session.ID(node.Id).Update(node)
			if err != nil {
				logger.Errorf("update node %v failed, %s", node, err)
				session.Rollback()
				return err
			}
		}
	}
	err := session.Commit()
	logger.Infof("commit nodes elasped %s", time.Since(start))
	return err
}

func (meicai *Meicai) initNodeHosts(url string, nodeStr string, nid int64) error {
	// TODO 请求失败，如何处理 ？
	hosts, err := QueryResourceByNode(url, meicai.Timeout, nodeStr, CmdbSourceHost)
	if err != nil {
		logger.Errorf("query resouce %s %s failed, %s", nodeStr, CmdbSourceHost, err)
		return err
	}

	return meicai.commitEndpoints(hosts, nid)
}

func (meicai *Meicai) initNodeNetworks(url string, nodeStr string, nid int64) error {
	// TODO 请求失败，如何处理 ？
	networks, err := QueryResourceByNode(url, meicai.Timeout, nodeStr, CmdbSourceNet)
	if err != nil {
		logger.Errorf("query resouce %s %s failed, %s", nodeStr, CmdbSourceNet, err)
		return err
	}

	return meicai.commitEndpoints(networks, nid)
}

func (meicai *Meicai) initNodeInstances(url string, nodeStr string, nid int64, apps map[string]*App) error {
	// TODO 请求失败，如何处理 ？
	instances, err := QueryResourceByNode(url, meicai.Timeout, nodeStr, CmdbSourceInst)
	if err != nil {
		logger.Errorf("query resouce %s %s failed, %s", nodeStr, CmdbSourceInst, err)
		return err
	}

	// 补充instance信息（app：basic）
	for _, instance := range instances {
		tags, err := dataobj2.SplitTagsString(instance.Tags)
		if err != nil {
			logger.Errorf("split instance tags %s failed, %s", instance.Tags, err)
		}
		if appcode, exist := tags["app"]; exist {
			if app, ok := apps[appcode]; ok {
				tags["basic"] = strconv.FormatBool(app.Basic)
			}
		}
	}
	return nil
	//return meicai.commitEndpoints(instances, nid)
}

func (meicai *Meicai) commitEndpoints(endpoints []*dataobj.Endpoint, nid int64) error {
	start := time.Now()
	session := meicai.DB["mon"].NewSession()
	defer session.Close()

	for _, host := range endpoints {
		// endpoint
		has, err := session.Exist(&dataobj.Endpoint{Id: host.Id})
		if err != nil || !has {
			logger.Infof("insert nid %d host %v", nid, host)
			_, err = session.Insert(host)
			if err != nil {
				logger.Errorf("insert endpoint %v failed, %s", host, err)
				session.Rollback()
				return err
			}
		} else {
			logger.Infof("update nid %d host %v", nid, host)
			_, err = session.ID(host.Id).Update(host)
			if err != nil {
				logger.Errorf("update endpoint %v failed, %s", host, err)
				session.Rollback()
				return err
			}
		}

		// node - endpoint
		nodeEndpoint := &NodeEndpoint{
			NodeId:     nid,
			EndpointId: host.Id,
		}
		exist, err := session.Exist(nodeEndpoint)
		if err != nil || !exist {
			logger.Infof("insert %v", nodeEndpoint)
			_, err = session.Insert(nodeEndpoint)
			if err != nil {
				logger.Errorf("insert node-endpoint %v failed, %s", nodeEndpoint, err)
				session.Rollback()
				return err
			}
		} else {
			logger.Infof("exist %v", nodeEndpoint)
		}

	}
	err := session.Commit()
	logger.Infof("commit endpoints elapsed %s", time.Since(start))
	return err
}
