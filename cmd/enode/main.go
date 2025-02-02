package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"

	_ "github.com/EducationEKT/EKT/api"
	"github.com/EducationEKT/EKT/conf"
	"github.com/EducationEKT/EKT/db"
	"github.com/EducationEKT/EKT/log"
	"github.com/EducationEKT/EKT/node"
	"github.com/EducationEKT/EKT/param"

	"github.com/EducationEKT/xserver/x_http"
)

const (
	version = "0.1"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU() - 1)
	var (
		help bool
		ver  bool
		m    string
		cfg  string
	)
	flag.BoolVar(&help, "h", false, "this help")
	flag.BoolVar(&ver, "v", false, "show version and exit")
	flag.StringVar(&m, "m", "adaptive", "specific node node: full sync OR fast sync OR delegate")
	flag.StringVar(&cfg, "c", "genesis.json", "set genesis.json file and start")
	flag.Parse()

	if help {
		flag.Usage()
		os.Exit(0)
	}

	if ver {
		fmt.Println(version)
		os.Exit(0)
	}

	err := InitService(cfg)
	if err != nil {
		fmt.Printf("Init service failed, %v \n", err)
		os.Exit(-1)
	}

	go node.Init(m)

	http.HandleFunc("/", x_http.Service)
}

func main() {
	fmt.Printf("server listen on :%d \n", conf.EKTConfig.Node.Port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", conf.EKTConfig.Node.Port), nil)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func InitService(confPath string) error {
	// init config
	// 初始化配置文件
	err := initConfig(confPath)
	if err != nil {
		return err
	}

	// init log service
	// 初始化日志服务
	err = initLog()
	if err != nil {
		return err
	}

	// init database service
	// 初始化levelDB服务
	initDB()

	// 初始化节点信息，包括私钥和peerId
	err = initPeerId()
	if err != nil {
		return err
	}

	// 初始化委托人节点
	param.InitBootNodes()

	return nil
}

func initPeerId() error {
	if len(conf.EKTConfig.PrivateKey) > 0 {
		log.Info("Peer private key is: %s ", conf.EKTConfig.PrivateKey)
		log.Info("Current peerId is: %s ", conf.EKTConfig.Node.Account)
	} else {
		log.Info("This is not delegate node.")
	}
	return nil
}

func initConfig(confPath string) error {
	return conf.InitConfig(confPath)
}

func initDB() {
	db.InitEKTDB(conf.EKTConfig.DBPath)
}

func initLog() error {
	log.InitLog(conf.EKTConfig.LogPath)
	return nil
}
