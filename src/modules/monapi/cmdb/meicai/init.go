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
	url := fmt.Sprintf("%s%s", meicai.OpsAddr, OpsApiResourcerPath)
	for _, node := range nodes {
		//初始化叶子节点的资源
		if node.Leaf == 1 && node.Note != "buffer" {
			logger.Infof("init leaf node endpoint, id=%d path=%s", node.Id, node.Path)
			nodeStr := node.Path
			// 主机资源
			if err := meicai.initNodeHosts(url, nodeStr, node.Id); err != nil {
				logger.Errorf("init node %s hosts failed, %s ", nodeStr, err)
			}
			// 网络资源
			if err := meicai.initNodeNetworks(url, nodeStr, node.Id); err != nil {
				logger.Errorf("init node %s network failed, %s", nodeStr, err)
			}

			// instance & app
			apps, err := QueryAppByNode(url, meicai.Timeout, nodeStr)
			if err == nil {
				appMap := make(map[string]*App, 0)
				for _, app := range apps {
					appMap[app.Code] = app
				}
				// 实例资源
				if err := meicai.initNodeAppInstances(url, nodeStr, node.Id, appMap); err != nil {
					logger.Errorf("init node %s app-instance failed, %s", nodeStr, err)
				}
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
			if _, err := session.Insert(node); err != nil {
				logger.Errorf("insert node %v failed, %s", node, err)
				_ = session.Rollback()
				return err
			}
		} else {
			logger.Infof("update node %v", node)
			if _, err := session.ID(node.Id).Update(node); err != nil {
				logger.Errorf("update node %v failed, %s", node, err)
				_ = session.Rollback()
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

func (meicai *Meicai) initNodeAppInstances(url string, nodeStr string, nid int64, apps map[string]*App) error {
	// TODO 请求失败，如何处理 ？
	appInstances, err := QueryAppInstanceByNode(url, meicai.Timeout, nodeStr, CmdbSourceInst)
	if err != nil {
		logger.Errorf("query resouce %s %s failed, %s", nodeStr, CmdbSourceInst, err)
		return err
	}

	// 补充instance信息（app：basic）
	for _, instance := range appInstances {
		tags, err := dataobj2.SplitTagsString(instance.Tags)
		if err != nil {
			logger.Errorf("split instance tags %s failed, %s", instance.Tags, err)
		}
		if app, ok := apps[instance.App]; ok {
			tags["basic"] = strconv.FormatBool(app.Basic)
		} else {
			tags["basic"] = strconv.FormatBool(false)
		}
		instance.Tags = dataobj2.SortedTags(tags)
		instance.NodeId = nid
	}
	return meicai.commitAppInstances(appInstances, nid)
}

func (meicai *Meicai) commitAppInstances(appInstances []*dataobj.AppInstance, nid int64) error {
	start := time.Now()
	session := meicai.DB["mon"].NewSession()
	defer session.Close()

	for _, instance := range appInstances {
		// app instance
		instModel := &dataobj.AppInstance{Uuid: instance.Uuid}
		has, err := session.Table("app_instance").Get(instModel)
		if err != nil || !has {
			logger.Infof("insert nid %d app-instance %v", nid, instance)
			if _, err := session.Table("app_instance").Insert(instance); err != nil {
				logger.Errorf("insert app-instance %v failed, %s", instance, err)
				_ = session.Rollback()
				return err
			}
		} else {
			logger.Infof("update nid %d app-instance %v", nid, instance)
			instance.Id = instModel.Id
			if _, err := session.Table("app_instance").ID(instance.Id).Update(instance); err != nil {
				logger.Errorf("update app-instance %v failed, %s", instance, err)
				_ = session.Rollback()
				return err
			}
		}
	}
	err := session.Commit()
	logger.Infof("commit app-instances elapsed %s", time.Since(start))
	return err
}

func (meicai *Meicai) commitEndpoints(endpoints []*dataobj.Endpoint, nid int64) error {
	start := time.Now()
	session := meicai.DB["mon"].NewSession()
	defer session.Close()

	for _, host := range endpoints {
		// endpoint
		endpointModel := &dataobj.Endpoint{Ident: host.Ident}
		has, err := session.Get(endpointModel)
		if err != nil || !has {
			logger.Infof("insert nid %d host %v", nid, host)
			if _, err := session.Insert(host); err != nil {
				logger.Errorf("insert endpoint %v failed, %s", host, err)
				_ = session.Rollback()
				return err
			}
		} else {
			logger.Infof("update nid %d host %v", nid, host)
			host.Id = endpointModel.Id
			if _, err := session.ID(host.Id).Update(host); err != nil {
				logger.Errorf("update endpoint %v failed, %s", host, err)
				_ = session.Rollback()
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
			if _, err := session.Insert(nodeEndpoint); err != nil {
				logger.Errorf("insert node-endpoint %v failed, %s", nodeEndpoint, err)
				_ = session.Rollback()
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
