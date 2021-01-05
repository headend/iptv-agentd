package main

import (
	"fmt"
	"github.com/headend/iptv-agentd/master"
	self_utils "github.com/headend/iptv-agentd/utils"
	"github.com/headend/share-module/configuration"
	"github.com/headend/share-module/configuration/static-config"
	"github.com/zhouhui8915/go-socket.io-client"
	"log"
	"sync"
	"time"
)


func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	// load config
	var conf  configuration.Conf
	conf.LoadConf()
	//log.Printf("%#v", conf)

	/*
	Xử lý thông tin kết nối
	Nếu thông tin không có trong config thì lấy từ static config
	*/
	gwHost := self_utils.GetGatewayHost(conf)
	gwPort := self_utils.GetGatewayPort(conf)
	// make authen params
	opts := &socketio_client.Options{
		Transport: static_config.GatewayTransportProtocol,
		Query:     make(map[string]string),
	}
	opts.Query["user"] = static_config.GatewayUser
	opts.Query["pwd"] = static_config.GatewayPassword
	var gatewayUrl string
	gatewayUrl = fmt.Sprintf("http://%s:%d/", gwHost, gwPort)
	var uri string
	uri = gatewayUrl + "socket.io/"
	//make channel control
	exitGWConnectChan := make(chan bool)
	exitControlChan := make(chan bool)
	profileRequestChan := make(chan string, 1)
	profileReceiveChan := make(chan string, 1)
	profileSignalReceiveChan := make(chan string, 1)
	profileVideoReceiveChan := make(chan string, 1)
	profileAudioReceiveChan := make(chan string, 1)
	profileChangeChan := make(chan string, 3)
	controlChan := make(chan string, 3)
	wg := new(sync.WaitGroup)
	wg.Add(3)

	// handle connect to gateway
	go func() {
		defer wg.Done()
		for {
			if master.RegisterGatewayClientSocket(uri,
				opts,
				exitGWConnectChan,
				profileRequestChan,
				profileReceiveChan,
				profileSignalReceiveChan,
				profileVideoReceiveChan,
				profileAudioReceiveChan,
				profileChangeChan,
				controlChan) {
				log.Printf("Wait for retry...")
				time.Sleep(10 * time.Second)
				continue
			} else {
				log.Printf("End...")
				return
			}
		}
	}()
	// Handle socket server to connect worker
	go func() {
		defer wg.Done()
		master.RegisterMasterSocket(wg,
			conf,
			profileRequestChan,
			profileChangeChan,
			profileReceiveChan,
			profileSignalReceiveChan,
			profileVideoReceiveChan,
			profileAudioReceiveChan)
	}()

	// Worker manager
	go func() {
		defer wg.Done()
		master.ControlWorkerHandle(exitControlChan,
			controlChan)
	}()

	wg.Wait()
	log.Println("Goodbye!")
}

//=========================================================================


