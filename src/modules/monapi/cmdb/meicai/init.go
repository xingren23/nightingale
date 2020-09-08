package meicai

import (
	"fmt"
	"log"
	"math/rand"
	"path"
	"strconv"
	"time"

	"github.com/didi/nightingale/src/modules/monapi/redisc"
	"github.com/didi/nightingale/src/toolkits/identity"

	"github.com/didi/nightingale/src/toolkits/stats"

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
	lockName := "monapi_sync_ops_lock"
	syncIntervalSecs := 30 * 60

	for {
		// The command returns -2 if the key does not exist.
		// The command returns -1 if the key exists but has no associated expire.
		ttl := redisc.TTL(lockName)
		if ttl != -2 {
			if ttl == -1 {
				if err := redisc.EXPIRE(lockName, syncIntervalSecs); err != nil {
					logger.Errorf("expire %s %d failed, %s", lockName, syncIntervalSecs, err)
				}
			}
			logger.Infof("get redis lock %s failed, ttl %d", lockName, ttl)
		} else {
			ok, err := redisc.SETNX(lockName, identity.Identity, syncIntervalSecs)
			if ok {
				err := meicai.SyncOps()
				if err != nil {
					stats.Counter.Set("cmdb.meicai.sync.err", 1)
					logger.Errorf("sync meicai node failed, %s", err)
				}
				stats.Counter.Set("cmdb.meicai.sync.count", 1)
			} else {
				logger.Errorf("'set %s %s nx ex %d' not ok, %s",
					lockName, identity.Identity, syncIntervalSecs, err)
			}
		}

		time.Sleep(5*time.Minute + time.Duration(rand.Int()%10)*time.Second)
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

	url := fmt.Sprintf("%s%s", meicai.OpsAddr, OpsApiResourcerPath)

	// 获取app
	appMap := make(map[string]*App, 0)
	apps, err := QueryAppByNode(url, meicai.Timeout, "corp.spruce")
	if err == nil {
		for _, app := range apps {
			appMap[app.Code] = app
		}
	} else {
		logger.Errorf("get apps failed, %s", err)
		return err
	}
	logger.Debugf("get apps %s", appMap)

	// 遍历节点
	for _, node := range nodes {
		//初始化叶子节点的资源
		if node.Leaf == 1 && node.Note != "buffer" {
			logger.Infof("init leaf node endpoint, id=%d path=%s", node.Id, node.Path)
			nodeStr := node.Path
			// endpoint: 主机&网络资源
			if err := meicai.initNodeEndpoints(url, nodeStr, node.Id); err != nil {
				logger.Errorf("init node %s hosts failed, %s ", nodeStr, err)
			}

			// appInstance: 实例资源
			if err := meicai.initNodeAppInstances(url, nodeStr, node.Id, appMap); err != nil {
				logger.Errorf("init node %s app-instance failed, %s", nodeStr, err)
			}

			time.Sleep(time.Duration(time.Millisecond) * 1)
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

	oldNodes, err := meicai.NodeGets()
	if err != nil {
		logger.Error("get nodes failed, error %s", err)
		return err
	}

	// insert or update node
	nodeMap := make(map[int64]*dataobj.Node)
	for _, node := range nodes {
		nodeMap[node.Id] = node
		has, err := session.Exist(&dataobj.Node{Id: node.Id})
		if err != nil || !has {
			logger.Infof("insert node %v", node)
			if _, err := session.Insert(node); err != nil {
				logger.Errorf("insert node %v failed, %s", node, err)
				_ = session.Rollback()
				stats.Counter.Set("cmdb.meicai.node.err", 1)
				return err
			}
			stats.Counter.Set("cmdb.meicai.node.insert", 1)
		} else {
			logger.Infof("update node %v", node)
			if _, err := session.ID(node.Id).Update(node); err != nil {
				logger.Errorf("update node %v failed, %s", node, err)
				_ = session.Rollback()
				stats.Counter.Set("cmdb.meicai.node.err", 1)
				return err
			}
			stats.Counter.Set("cmdb.meicai.node.update", 1)
		}
	}

	// delete old node
	for _, old := range oldNodes {
		if _, exist := nodeMap[old.Id]; !exist {
			logger.Infof("delete node %v", old)
			session.Delete(old)
			stats.Counter.Set("cmdb.meicai.node.delete", 1)
		}
	}

	err = session.Commit()
	logger.Infof("commit nodes elasped %s", time.Since(start))
	return err
}

func (meicai *Meicai) initNodeEndpoints(url string, nodeStr string, nid int64) error {
	oldEndpoints, err := meicai.EndpointUnderLeafs([]int64{nid})
	if err != nil {
		logger.Errorf("get endpointIds by node id %d failed, error %s", nid, err)
		return err
	}

	// TODO 请求失败，如何处理 ？
	networks, err := QueryResourceByNode(url, meicai.Timeout, nodeStr, CmdbSourceNet)
	if err != nil {
		logger.Errorf("query network resource %s %s failed, %s", nodeStr, CmdbSourceNet, err)
		return err
	}
	hosts, err := QueryResourceByNode(url, meicai.Timeout, nodeStr, CmdbSourceHost)
	if err != nil {
		logger.Errorf("query host resource %s %s failed, %s", nodeStr, CmdbSourceHost, err)
		return err
	}

	endpoints := append(networks, hosts...)

	return meicai.commitEndpoints(endpoints, oldEndpoints, nid)
}

func (meicai *Meicai) initNodeAppInstances(url string, nodeStr string, nid int64, apps map[string]*App) error {
	oldInstances, err := meicai.AppInstanceUnderLeafs([]int64{nid})
	if err != nil {
		logger.Errorf("get appInstances under node %d failed, error %s", nid, err)
		return err
	}

	// TODO 请求失败，如何处理 ？
	appInstances, err := QueryAppInstanceByNode(url, meicai.Timeout, nodeStr, CmdbSourceInst)
	if err != nil {
		logger.Errorf("query resource %s %s failed, %s", nodeStr, CmdbSourceInst, err)
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

	return meicai.commitAppInstances(appInstances, oldInstances, nid)
}

func (meicai *Meicai) commitAppInstances(appInstances []*dataobj.AppInstance, oldInstances []dataobj.AppInstance,
	nid int64) error {
	start := time.Now()
	session := meicai.DB["mon"].NewSession()
	defer session.Close()

	instanceMap := make(map[string]*dataobj.AppInstance)
	// insert or update appInstance
	for _, instance := range appInstances {
		instanceMap[instance.Uuid] = instance
		instModel := &dataobj.AppInstance{Uuid: instance.Uuid}
		has, err := session.Table("app_instance").Get(instModel)
		if err != nil || !has {
			logger.Infof("insert nid %d app-instance %v", nid, instance)
			if _, err := session.Table("app_instance").Insert(instance); err != nil {
				logger.Errorf("insert app-instance %v failed, %s", instance, err)
				_ = session.Rollback()
				stats.Counter.Set("cmdb.meicai.appinstance.err", 1)

				return err
			}
			stats.Counter.Set("cmdb.meicai.appinstance.insert", 1)
		} else {
			logger.Debugf("update nid %d app-instance %v", nid, instance)
			instance.Id = instModel.Id
			if _, err := session.Table("app_instance").ID(instance.Id).Update(instance); err != nil {
				logger.Errorf("update app-instance %v failed, %s", instance, err)
				_ = session.Rollback()
				stats.Counter.Set("cmdb.meicai.appinstance.err", 1)
				return err
			}
			stats.Counter.Set("cmdb.meicai.appinstance.update", 1)
		}
	}

	// delete appInstance
	for _, old := range oldInstances {
		if _, exist := instanceMap[old.Uuid]; !exist {
			logger.Infof("delete nid %d app-instance %v", nid, old)
			session.Delete(old)
			stats.Counter.Set("cmdb.meicai.appinstance.delete", 1)
		}
	}

	err := session.Commit()
	logger.Infof("commit app-instances elapsed %s", time.Since(start))
	return err
}

func (meicai *Meicai) commitEndpoints(endpoints []*dataobj.Endpoint, oldEndpoints []dataobj.Endpoint,
	nid int64) error {
	start := time.Now()
	session := meicai.DB["mon"].NewSession()
	defer session.Close()

	endpointMap := make(map[string]*dataobj.Endpoint)
	// insert or update endpoint and node-endpoint
	for _, host := range endpoints {
		endpointMap[host.Ident] = host

		// endpoint
		endpointModel := &dataobj.Endpoint{Ident: host.Ident}
		has, err := session.Get(endpointModel)
		if err != nil || !has {
			logger.Infof("insert nid %d host %v", nid, host)
			if _, err := session.Insert(host); err != nil {
				logger.Errorf("insert endpoint %v failed, %s", host, err)
				_ = session.Rollback()
				stats.Counter.Set("cmdb.meicai.endpoint.err", 1)
				return err
			}
			stats.Counter.Set("cmdb.meicai.endpoint.insert", 1)
		} else {
			logger.Debugf("update nid %d host %v", nid, host)
			host.Id = endpointModel.Id
			if _, err := session.ID(host.Id).Update(host); err != nil {
				logger.Errorf("update endpoint %v failed, %s", host, err)
				_ = session.Rollback()
				stats.Counter.Set("cmdb.meicai.endpoint.err", 1)
				return err
			}
			stats.Counter.Set("cmdb.meicai.endpoint.update", 1)
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
				stats.Counter.Set("cmdb.meicai.node.endpoint.err", 1)
				return err
			}
			stats.Counter.Set("cmdb.meicai.node.endpoint.insert", 1)
		} else {
			logger.Debugf("exist %v", nodeEndpoint)
		}
	}

	// delete endpoint and node-endpoint
	for _, old := range oldEndpoints {
		if _, exist := endpointMap[old.Ident]; !exist {
			logger.Infof("delete endpoint %v", old)
			session.Delete(old)
			stats.Counter.Set("cmdb.meicai.endpoint.delete", 1)

			nodeEndpoint := &NodeEndpoint{
				NodeId:     nid,
				EndpointId: old.Id,
			}
			logger.Infof("delete node-endpoint %v", nodeEndpoint)
			session.Delete(nodeEndpoint)
			stats.Counter.Set("cmdb.meicai.node.endpoint.delete", 1)
		}
	}

	err := session.Commit()
	logger.Infof("commit endpoints elapsed %s", time.Since(start))
	return err
}
