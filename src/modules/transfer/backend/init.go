package backend

type InfluxdbSection struct {
	Enabled   bool   `yaml:"enabled"`
	Name      string `yaml:"name"`
	Batch     int    `yaml:"batch"`
	MaxRetry  int    `yaml:"maxRetry"`
	WorkerNum int    `yaml:"workerNum"`
	Timeout   int    `yaml:"timeout"`
	Address   string `yaml:"address"`
	Database  string `yaml:"database"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	Precision string `yaml:"precision"`
}

type OpenTsdbSection struct {
	Enabled     bool   `yaml:"enabled"`
	Name        string `yaml:"name"`
	Batch       int    `yaml:"batch"`
	ConnTimeout int    `yaml:"connTimeout"`
	CallTimeout int    `yaml:"callTimeout"`
	WorkerNum   int    `yaml:"workerNum"`
	MaxConns    int    `yaml:"maxConns"`
	MaxIdle     int    `yaml:"maxIdle"`
	MaxRetry    int    `yaml:"maxRetry"`
	Address     string `yaml:"address"`
}

type KafkaSection struct {
	Enabled      bool   `yaml:"enabled"`
	Name         string `yaml:"name"`
	Topic        string `yaml:"topic"`
	BrokersPeers string `yaml:"brokersPeers"`
	ConnTimeout  int    `yaml:"connTimeout"`
	CallTimeout  int    `yaml:"callTimeout"`
	MaxRetry     int    `yaml:"maxRetry"`
	KeepAlive    int64  `yaml:"keepAlive"`
	SaslUser     string `yaml:"saslUser"`
	SaslPasswd   string `yaml:"saslPasswd"`
}

type ClusterNode struct {
	Addrs []string `json:"addrs"`
}

type TsdbSection struct {
	Enabled      bool   `yaml:"enabled"`
	Name         string `yaml:"name"`
	Batch        int    `yaml:"batch"`
	ConnTimeout  int    `yaml:"connTimeout"`
	CallTimeout  int    `yaml:"callTimeout"`
	WorkerNum    int    `yaml:"workerNum"`
	MaxConns     int    `yaml:"maxConns"`
	MaxIdle      int    `yaml:"maxIdle"`
	IndexTimeout int    `yaml:"indexTimeout"`

	Replicas    int                     `yaml:"replicas"`
	Cluster     map[string]string       `yaml:"cluster"`
	ClusterList map[string]*ClusterNode `json:"clusterList"`
}

type JudgeSection struct {
	Batch       int    `yaml:"batch"`
	ConnTimeout int    `yaml:"connTimeout"`
	CallTimeout int    `yaml:"callTimeout"`
	WorkerNum   int    `yaml:"workerNum"`
	MaxConns    int    `yaml:"maxConns"`
	MaxIdle     int    `yaml:"maxIdle"`
	HbsMod      string `yaml:"hbsMod"`
}

type BackendSection struct {
	Storage  string `yaml:"storage"`
	StraPath string `yaml:"straPath"`

	Judge    JudgeSection    `yaml:"judge"`
	Tsdb     TsdbSection     `yaml:"tsdb"`
	Influxdb InfluxdbSection `yaml:"influxdb"`
	OpenTsdb OpenTsdbSection `yaml:"opentsdb"`
	Kafka    KafkaSection    `yaml:"kafka"`
}

var (
	defaultStorage    string
	StraPath          string
	tsdbStorage       *TSDBStorage
	openTSDBStorage   *OpenTSDBStorage
	influxDBStorage   *InfluxDBStorage
	kafkaPushEndpoint *KafkaPushEndpoint
)

func Init(cfg BackendSection) {
	defaultStorage = cfg.Storage
	StraPath = cfg.StraPath

	// init judge
	InitJudge(cfg.Judge)

	// init tsdb storage
	if cfg.Tsdb.Enabled {
		tsdbStorage = &TSDBStorage{
			section: cfg.Tsdb,
		}
		tsdbStorage.Init()
	}

	// init influxdb storage
	if cfg.Influxdb.Enabled {
		influxDBStorage = &InfluxDBStorage{
			section: cfg.Influxdb,
		}
		influxDBStorage.Init()

	}
	// init opentsdb storage
	if cfg.OpenTsdb.Enabled {
		openTSDBStorage = &OpenTSDBStorage{
			section: cfg.OpenTsdb,
		}
		openTSDBStorage.Init()
	}
	// init kafka
	if cfg.Kafka.Enabled {
		kafkaPushEndpoint = &KafkaPushEndpoint{
			section: cfg.Kafka,
		}
		kafkaPushEndpoint.Init()
	}
}
