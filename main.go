package main

import (
	"encoding/json"
	"fmt"
	socketio "github.com/googollee/go-socket.io"
	"github.com/headend/iptv-agentd/master"
	"github.com/headend/share-module/configuration"
	"github.com/headend/share-module/configuration/socket-event"
	"github.com/headend/share-module/configuration/static-config"
	file_and_directory "github.com/headend/share-module/file-and-directory"
	"github.com/headend/share-module/model"
	agentModel "github.com/headend/share-module/model/agentd"
	"github.com/headend/share-module/shellout"
	"github.com/zhouhui8915/go-socket.io-client"
	"log"
	"strconv"
	"sync"
	"syscall"
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
	var gwHost string
	if conf.AgentGateway.Gateway != "" {
		gwHost = conf.AgentGateway.Gateway
	} else {
		if conf.AgentGateway.Host != "" {
			gwHost = conf.AgentGateway.Host
		} else {
			gwHost = static_config.GatewayHost
		}
	}
	var gwPort uint16
	if conf.AgentGateway.Port != 0 {
		gwPort = conf.AgentGateway.Port
	} else {
		gwPort = static_config.GatewayPort
	}
	// make authen params
	opts := &socketio_client.Options{
		Transport: static_config.GatewayTransportProtocol,
		Query:     make(map[string]string),
	}
	opts.Query["user"] = static_config.GatewayUser
	opts.Query["pwd"] = static_config.GatewayPassword
	var gatewayUrl string
	gatewayUrl = fmt.Sprintf("http://%s:%d/", gwHost, gwPort)
	//gatewayUrl = "http://" + gwHost + ":" + string(gwPort) + "/"
	var uri string
	uri = gatewayUrl+ "socket.io/"
	//make channel control
	chDestroyMaster := make(chan bool)
	chDestroyInternalSocket := make(chan bool)
	var client *socketio_client.Client
	var err error

	// make socket
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	// connect to gateway
	for {
		select {
		case <-chDestroyMaster:
			log.Println("[Agentd] Reconnect...")
			time.Sleep(10 * time.Second)
			continue
		default:
			var wg sync.WaitGroup
			wg.Add(1)
			client, err = socketio_client.NewClient(uri, opts)
			if err != nil {
				log.Printf("[Agentd] NewClient error:%v\n", err)
				log.Println("[Agentd] Reconnect...")
				wg.Done()
				time.Sleep(10 * time.Second)
				continue
			}

			client.On(socket_event.Loi, func() {
				log.Printf("[Agentd] on error\n")
			})
			client.On(socket_event.KetNoi, func(msg string) {
				log.Printf("[Agentd] Connected whith message: %v\n", msg)

			})
			client.On(socket_event.NhanFile, func(msg string) {
				master.MasterReceiveFile(msg, gatewayUrl, client)
			})

			client.On(socket_event.TinNhan, func(msg string) {
				log.Printf("[Agentd] on message:%v\n", msg)
			})
			client.On(socket_event.NgatKetNoi, func() {
				log.Printf("[Agentd] Disconnect from server")
				chDestroyInternalSocket <- true
				chDestroyMaster <- true
				wg.Done()
			})

			client.On(socket_event.ThucThiLenh, func(msg string) {
				master.OnExecuteRequestHandle(msg, client, server)
			})
			client.On(socket_event.DieuKhien, func(msg string) {
				log.Println(msg)
				var ctlRequestData *model.AgentCTLQueueRequest
				json.Unmarshal([]byte(msg), &ctlRequestData)
				go func() {
					var runThread int
					if ctlRequestData.RunThread != 0 {
						runThread = ctlRequestData.RunThread
					} else {
						runThread = 1
					}
					runThreadString := fmt.Sprintf("%d", runThread)
					switch ctlRequestData.ControlType {
					case static_config.StartMonitorSignal:
						log.Println("[Agentd] run signal worker")
						appToRUn := fmt.Sprintf("%s/iptv-agentd", static_config.BinaryPath)
						err, exitCode, stdout, stderr := shellout.RunExternalCmd(appToRUn, []string{"-m", "daemon", "-t", "signal", "-n", runThreadString}, 0)
						log.Printf("err: %s", err.Error())
						log.Printf("exitCode: %d", exitCode)
						log.Printf("stdout: %s", stdout)
						log.Printf("stderr: %s", stderr)
					case static_config.StartMonitorVideo:
						log.Println("[Agentd] run signal worker")
						appToRUn := fmt.Sprintf("%s/iptv-agentd", static_config.BinaryPath)
						err, exitCode, stdout, stderr := shellout.RunExternalCmd(appToRUn, []string{"-m", "daemon", "-t", "video", "-n", runThreadString}, 0)
						log.Printf("err: %s", err.Error())
						log.Printf("exitCode: %d", exitCode)
						log.Printf("stdout: %s", stdout)
						log.Printf("stderr: %s", stderr)
					case static_config.StartMonitorAudio:
						log.Println("[Agentd] run signal worker")
						appToRUn := fmt.Sprintf("%s/iptv-agentd", static_config.BinaryPath)
						err, exitCode, stdout, stderr := shellout.RunExternalCmd(appToRUn, []string{"-m", "daemon", "-t", "audio", "-n", runThreadString}, 0)
						log.Printf("err: %s", err.Error())
						log.Printf("exitCode: %d", exitCode)
						log.Printf("stdout: %s", stdout)
						log.Printf("stderr: %s", stderr)
					case static_config.StopMonitorSignal:
						pidFilePath := fmt.Sprintf("%s/run/signal.pid", static_config.InstallationPath)
						var pidFile file_and_directory.MyFile
						pidFile.Path = pidFilePath
						pidString,_ := pidFile.Read()
						pid, _ := strconv.Atoi(pidString)
						err = syscall.Kill(pid, 15)
						if err != nil {
							log.Println("Success stop signal monitor")
						}
					case static_config.StopMonitorVideo:
						pidFilePath := fmt.Sprintf("%s/run/video.pid", static_config.InstallationPath)
						var pidFile file_and_directory.MyFile
						pidFile.Path = pidFilePath
						pidString,_ := pidFile.Read()
						pid, _ := strconv.Atoi(pidString)
						err = syscall.Kill(pid, 15)
						if err != nil {
							log.Println("Success stop video monitor")
						}
					case static_config.StopMonitorAudio:
						pidFilePath := fmt.Sprintf("%s/run/audio.pid", static_config.InstallationPath)
						var pidFile file_and_directory.MyFile
						pidFile.Path = pidFilePath
						pidString,_ := pidFile.Read()
						pid, _ := strconv.Atoi(pidString)
						err = syscall.Kill(pid, 15)
						if err != nil {
							log.Println("Success stop audio monitor")
						}
					default:
						log.Println("Not support")
					}
				}()

			})

			client.On("profile-monitor-response", func(msg string) {

				// Transfer message to worker
				var data agentModel.MonitorInputForAgent
				data.LoadFromJsonString(msg)
				switch data.MonitorType {
				case static_config.Video:
					server.BroadcastToRoom("/", "video", "profile-monitor-response", msg)
				case static_config.Audio:
					server.BroadcastToRoom("/", "audio", "profile-monitor-response", msg)
				case static_config.Signal:
					server.BroadcastToRoom("/", "signal", "profile-monitor-response", msg)
				default:
					server.BroadcastToRoom("/", socket_event.NhomChung, "profile-monitor-response", msg)
				}

			})
			// Then register socket to listen from worker
			log.Println("[Agentd] Start internal socket")
			master.RegisterMasterSocket(client, server, conf, &chDestroyInternalSocket)
			log.Println("Wait event from client")
			wg.Wait()
		}
	}
	log.Println("Goodbye!")
}


