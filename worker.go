package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/headend/iptv-agentd/utils"
	self_utils "github.com/headend/iptv-agentd/utils"
	selfWorker "github.com/headend/iptv-agentd/worker"
	"github.com/headend/share-module/configuration"
	socket_event "github.com/headend/share-module/configuration/socket-event"
	static_config "github.com/headend/share-module/configuration/static-config"
	"github.com/headend/share-module/file-and-directory"
	model "github.com/headend/share-module/model/agentd"
	"github.com/headend/share-module/shellout"
	socketio_client "github.com/zhouhui8915/go-socket.io-client"
	selfutils "github.com/headend/iptv-agentd/utils"
	"os"
	"strconv"
	"sync"
	"time"

	//"github.com/headend/iptv-agentd/worker/event-handle"
	"log"
)

type AgentdPingResponse struct {
	MasterVersion	string
	WorkerVersion	string
	WorkerTheard	string
	CpuCore			int
	CpuUsage		float32
	CpuUtil			float32
	CpuLoad			float32
	Ram				int
	RamUsage		int
	Ip 				string
}

func main()  {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	confFilePtr := flag.String("c", static_config.ConfigFilePath, "Configure file")
	modePtr := flag.String("m", "daemon", "Run mode: daemon(default)/urgent")
	sourcePtr := flag.String("s", "", "Source ip multicast (required) if urgent mode")
	workerNumPtr := flag.String("n", "1", "Concurrency worker")
	monitorTypePtr := flag.String("t", "signal", "monitor type: signal/video/audio")
	flag.Parse()
	// load config
	var conf configuration.Conf
	if confFilePtr != nil {
		conf.ConfigureFile = *confFilePtr
	}
	var runMode string
	if *modePtr != "" {
		runMode = *modePtr
		if *modePtr != "daemon" && *modePtr != "urgent" {
			log.Fatalf("Run mode: daemon(default)/urgent. Not support your mode %s", *modePtr)
		}
	}

	var threadNum int
	if *workerNumPtr != "" {
		threadNum, err := strconv.Atoi(*workerNumPtr)
		if err != nil {
			log.Fatal(err)
		} else {
			log.Printf("Worker run with %d thread(s)\n", threadNum)
		}
	}
	var moitorType int
	selfPid := os.Getpid()
	SelfPidString := fmt.Sprintf("%d", selfPid)
	var pidFile file_and_directory.MyFile
	var pidPathFile string
	logFilePath := fmt.Sprintf("/opt/iptv/logs/%s.log", *monitorTypePtr)
	switch *monitorTypePtr {
	case "video":
		moitorType = static_config.Video
		pidPathFile = fmt.Sprintf("%s/run/video.pid", static_config.InstallationPath)
		pidFile.Path = pidPathFile
		pidFile.WriteString(SelfPidString)
	case "audio":
		moitorType = static_config.Audio
		pidPathFile = fmt.Sprintf("%s/run/audio.pid", static_config.InstallationPath)
		pidFile.Path = pidPathFile
		pidFile.WriteString(SelfPidString)
	default:
		moitorType = static_config.Signal
		pidPathFile = fmt.Sprintf("%s/run/signal.pid", static_config.InstallationPath)
		pidFile.Path = pidPathFile
		pidFile.WriteString(SelfPidString)
	}

	f, err := os.OpenFile(logFilePath, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)

	conf.LoadConf()


	masterHost, masterPort := utils.GetMasterConnectionInfo(conf)

	// make authen params
	opts := &socketio_client.Options{
		Transport: static_config.GatewayTransportProtocol,
		Query:     make(map[string]string),
	}
	opts.Query["user"] = static_config.MasterUser
	opts.Query["pwd"] = static_config.MasterPassword
	var gatewayUrl string
	gatewayUrl = fmt.Sprintf("http://%s:%d/", masterHost, masterPort)
	//gatewayUrl = "http://" + gwHost + ":" + string(gwPort) + "/"
	fmt.Printf("Conect to: %s \n", gatewayUrl)
	var uri string
	uri = gatewayUrl+ "socket.io/"


	//make channel control
	chDestroyWorker := make(chan bool)
	chQuitRequestProfile := make(chan bool)
	for  {
		select {
		case <-chDestroyWorker:
			break
		default:
			var tasks []*selfWorker.Task
			p := selfWorker.NewPool(tasks, threadNum)
			go p.Run()

			client, err := socketio_client.NewClient(uri, opts)
			var wg sync.WaitGroup
			if err != nil {
				log.Printf("NewClient error:%v\n", err)
				log.Printf("Wait for retry connect:%v\n", err)
				// Quit goroutine request profile
				chQuitRequestProfile <- true
				time.Sleep(10 * time.Second)
				continue
			} else {
				wg.Add(1)
			}
			//client.On(socket_event.DieuKhien, func(signal int) {
			//	switch signal {
			//	case static_config.StopAgentd :
			//		defer wg.Done()
			//		log.Printf("Stoping... (System request)\n")
			//	default:
			//		log.Println("do not thing...")
			//	}
			//})

			client.On(socket_event.Loi, func() {
				log.Printf("on error\n")
				wg.Done()
			})
			client.On(socket_event.KetNoi, func(msg string) {
				log.Printf("Connected whith message: %v\n", msg)
				// register room
				moitorTypeString := fmt.Sprintf("%d", moitorType)
				client.Emit("register-monitor-type", moitorTypeString)
			})

			client.On(socket_event.TinNhan, func(msg string) {
				log.Printf("on message:%v\n", msg)
			})
			client.On(socket_event.NgatKetNoi, func() {
				log.Printf("Disconnect from server\nGoodbye!")
				wg.Done()
			})
			fmt.Println("WAITING spawn threat(s)...")
			if runMode == "daemon" {
				client.On(socket_event.DieuKhien, func(msg string) {
					log.Println("Receive terminate signal")
					ctlType, err := strconv.Atoi(msg)
					if err != nil {
						log.Println(err)
						return
					}
					switch ctlType {
					case static_config.StopWorker:
						// Send exit channel
						chDestroyWorker <- true
					default:
						log.Println("Worker not support")
					}
				})
				go func() {
					for {
						select {
						case <-chQuitRequestProfile:
							log.Println("Sync profile closed")
							break
						default:
							// Do other stuff
							if len(p.TasksChan) <= 3 {
								log.Println("request new profile")
								log.Println(len(p.TasksChan))
								msgToSend := fmt.Sprintf("%d", moitorType)
								errSendProfileRequest := client.Emit("profile-monitor-request", msgToSend)
								if errSendProfileRequest != nil {
									println(err)
								} else {
									chQuitRequestProfile <- true
								}
							} else {
								log.Printf("%d task(s) left in queue", len(p.TasksChan))
								//chQuitRequestProfile <- true
							}
							log.Println("Wait for sync profile")
							time.Sleep(10 * time.Second)
						}
					}
				}()
				client.On("ping", func(msg string) {
					var rspData AgentdPingResponse
					var rspMsg string
					tmpMsg, err := json.Marshal(rspData)
					if err != nil {
						fmt.Println(err)
						rspMsg = err.Error()
					} else {
						rspMsg = string(tmpMsg)
					}
					client.Emit("pong", rspMsg)
				})

				client.On(socket_event.ThucThiLenh, func(command string) {
					fmt.Printf(command)
					var msg string
					err, exitCode, stdout, stderr := shellout.ExecuteCommand("/bin/sh", "ls -al")
					if err != nil {
						msg += fmt.Sprintf("Fatal Error: %s", err.Error())
					} else {
						msg = fmt.Sprintf("Exit code: %d", exitCode)
						if stderr != "" {
							msg += fmt.Sprintf("Stderr: %s", stderr)
						} else {
							msg += fmt.Sprintf("Stdout: %s", stdout)
						}
						for i := 1; i <= 5; i++ {
							fmt.Println(i)
							time.Sleep(1 * time.Second)
						}

						client.Emit(socket_event.KetQuaThucThiLenh, msg)
					}
				})
				// Nhan profile list
				client.On("profile-monitor-response", func(msg string) {
					log.Println(msg)
					var monitorProfileData model.MonitorInputForAgent
					monitorProfileData.LoadFromJsonString(msg)
					log.Println(monitorProfileData)
					for _, profile := range monitorProfileData.ProfileList{
						log.Println("Do monitor")
						multicatsStream := fmt.Sprintf("%s:1234", profile.MulticastIP)
						_, checkcode := selfutils.CheckSourceMulticast(multicatsStream)
						if checkcode != profile.Status {
							time.Sleep(20*time.Second)
							//recheck
							err, checkcode = selfutils.CheckSourceMulticast(multicatsStream)
							if checkcode != profile.Status {
								log.Println("Wait for recheck")
								msg := fmt.Sprintf("Status has change from %d to %d\n", profile.Status, checkcode)
								log.Println(msg)
								var signalStatus bool
								if checkcode == 1 {
									signalStatus = true
								}
								changeStatusMessageData := model.ProfileChangeStatus{
									MonitorType:     moitorType,
									MonitorID:       profile.MonitorId,
									ProfileId:       profile.ProfileId,
									AgentId:         profile.AgentId,
									OldStatus:       profile.Status,
									NewStatus:       checkcode,
									OldSignalStatus: false,
									NewSignalStatus: signalStatus,
									OldVideoStatus:  false,
									NewVideoStatus:  false,
									OldAudioStatus:  false,
									NewAudioStatus:  false,
									EventTime:       time.Now().Unix(),
								}
								changeStatusMessageSting, _ := changeStatusMessageData.GetJsonString()
								err2 := client.Emit("monitor-response", changeStatusMessageSting)
								if err2 != nil {
									log.Println(err2)
								}
							}
						}
						log.Printf("Result status: %d", checkcode)
					}
					msgToSend := fmt.Sprintf("%d", moitorType)
					errSendProfileRequest := client.Emit("profile-monitor-request", msgToSend)
					if errSendProfileRequest != nil {
						log.Println(err)
					}
					//for _,profile := range monitorProfileData.ProfileList {
					//	newTask := selfWorker.NewTask(profile)
					//	log.Println(newTask)
					//	//log.Println(len(p.TasksChan))
					//	p.TasksChan <- newTask
					//	//log.Println(len(p.TasksChan))
				})
				//======================================================================================================
				// Run mode as Urgent
			} else {
				log.Printf("Worker runas %s mode \n", runMode)
				// check source ip
				if *sourcePtr != "" {
					sourceMulticast := *sourcePtr
					err, checkcode := self_utils.CheckSourceMulticast(sourceMulticast)
					if err != nil {
						// retry if not found source
						if err.Error() == "killed" && checkcode == 0 {
							log.Println("Wait for recheck.")
							time.Sleep(60 * time.Second)
							_, checkcode = self_utils.CheckSourceMulticast(sourceMulticast)
							SourceCheckingReport(checkcode)
						} else {
							log.Println(err)
						}
					} else {
						SourceCheckingReport(checkcode)
					}
				} else {
					log.Println("Urgent mode required source ip")
				}
				wg.Done()
			}
			//==========================================================================================================
			// End urgent mode

			wg.Wait()
			log.Println("Reconnect...")
			//
		}
	}
	log.Println("Worker done. Good bye!")
}

func SourceCheckingReport(checkcode int64) {
	switch checkcode {
	case static_config.SourceNotOK:
		log.Println("NotOK")
	case static_config.SourceOK:
		log.Println("OK")
	case static_config.SourceNoVideo:
		log.Println("No Video")
	case static_config.SourceNoAudio:
		log.Println("No Audio")
	default:
		log.Println("Unknow")
	}
}


