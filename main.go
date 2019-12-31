package main

import (
	"log"
	"fmt"
	"time"
	"net/http"
	"os"
	"os/signal"

	"github.com/hashicorp/consul/api"

	"github.com/bvs/utils"
	"github.com/bvs/routers"
	"github.com/bvs/zkservice"
)

type program struct {
	hostIp     string
	httpPort   int
	httpServer *http.Server
}

func (p *program) RegisterService() (err error) {
	
	consulSec, err := utils.Conf().GetSection("consul")
	if err != nil {
		log.Println(err)
		return err
	}
	consul_host := consulSec.Key("host_addr").MustString("127.0.0.1:8500")
	serverId := consulSec.Key("service_id").MustString("av-domain-service8080")
	serverName := consulSec.Key("service_name").MustString("av-domain-service")

	conf := api.DefaultConfig()
	conf.Address = consul_host

	// Create client
	client, err := api.NewClient(conf)
	if err != nil {
		log.Fatalf("err: %v", err)
		return err
	}

	consul_agent := client.Agent()

	check := &api.AgentServiceCheck{
		HTTP:     fmt.Sprintf("http://%s:%d/ping", p.hostIp, p.httpPort),
		Interval: "10s",
		Timeout:  "2s",
	}

	service1 := &api.AgentServiceRegistration{
		ID:      serverId,
		Name:    serverName,
		Tags:    []string{"123", "abc"},
		Port:    p.httpPort,
		Address: p.hostIp,
		Check:   check,
	}

	err = consul_agent.ServiceRegister(service1)
	if err != nil {
		log.Fatalf("reg service err: %s", err)
	}
	
	return err
}

func (p *program) StartHTTP() (err error) {
	p.httpServer = &http.Server{
		Addr:              fmt.Sprintf("%s:%d", p.hostIp, p.httpPort),
		Handler:           routers.Router,
		ReadHeaderTimeout: 5 * time.Second,
	}
	link := fmt.Sprintf("http://%s:%d", p.hostIp/*utils.LocalIP()*/, p.httpPort)
	log.Println("http server start -->", link)
	go func() {
		if err := p.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("start http server error", err)
		}
		log.Println("http server end")
	}()
	return
}

func main() {

	log.Println("********** START **********")

	hostIp := utils.Conf().Section("net").Key("local_ip").MustString("127.0.0.1")
	httpPort := utils.Conf().Section("http").Key("port").MustInt(8080)
	zkHostAddr := utils.Conf().Section("zk").Key("host_addr").MustString("127.0.0.1:2181")

	p := &program{
		hostIp:   hostIp,
		httpPort:   httpPort,
	}
	
	var err error
	if utils.IsPortInUse(p.httpPort) {
		log.Printf("HTTP port[%d] In Use\n", p.httpPort)
		return
	}

	log.SetOutput(utils.GetLogWriter())
	err = routers.Init()
	if err != nil {
		log.Println("routers init failed")
		return
	}

	zkservice.NewZkNodeManager([]string{zkHostAddr})
	go zkservice.NodeManager.Start()

	p.StartHTTP()
	p.RegisterService()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Kill, os.Interrupt)
	s := <-sigs
    log.Println("Exit get signal:", s)
}

