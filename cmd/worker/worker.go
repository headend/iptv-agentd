package main

import (
	"flag"
	"fmt"
	"github.com/headend/iptv-agentd/utils"
	self_utils "github.com/headend/iptv-agentd/utils"
	"github.com/headend/share-module/configuration"
	socket_event "github.com/headend/share-module/configuration/socket-event"
	static_config "github.com/headend/share-module/configuration/static-config"
	"github.com/headend/share-module/file-and-directory"
	model "github.com/headend/share-module/model/agentd"
	socketio_client "github.com/zhouhui8915/go-socket.io-client"
	"os"
	"strconv"
	"sync"
	"time"

	//"github.com/headend/iptv-agentd/worker/event-handle"
	"log"
)

const version = "1.5"

func main()  {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	confFilePtr := flag.String("c", static_config.ConfigFilePath, "Configure file")
	modePtr := flag.String("m", "daemon", "Run mode: daemon(default)/urgent")
	sourcePtr := flag.String("s", "", "Source ip multicast (required) if urgent mode")
	workerNumPtr := flag.String("n", "1", "Concurrency worker")
	monitorTypePtr := flag.String("t", "signal", "monitor type: signal/video/audio")
	versionPtr := flag.String("v", "", "Get version (anything value)")
	flag.Parse()
	// check version
	if *versionPtr != "" {
		fmt.Print(version)
		return
	}
	// load config
	var conf configuration.Conf
	if confFilePtr != nil {
		conf.ConfigureFile = *confFilePtr
	}
	runMode := GetRunmode(modePtr)

	threadNum := GetWorkerCurrency(workerNumPtr)
	log.Println(threadNum)
	moitorType, logFilePath := GetMonitorTypeAndRegisterPidFile(monitorTypePtr)
	println(logFilePath)
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
	fmt.Printf("Conect to: %s \n", gatewayUrl)
	var uri string
	uri = gatewayUrl + "socket.io/"

	//make channel control
	exitMasterConChan := make(chan bool)
	profileChangeChan := make(chan string)
	profileRequestChan := make(chan string)
	profileReceiveChan := make(chan string)

	var wg sync.WaitGroup
	wg.Add(1)

	//====================================
	// Register socket client connect to master
	go func() {
		defer wg.Done()
		for {
			masterConnOk := RegisterMasterConnect(uri, opts, exitMasterConChan, moitorType, profileReceiveChan, profileChangeChan, profileRequestChan)
			if masterConnOk {
				continue
			} else {
				log.Printf("Wait for retry...")
				time.Sleep(10*time.Second)
				continue
			}
		}
	}()
	// run check mode
	if runMode == "ugrent" {
		log.Printf("Worker runas %s mode \n", runMode)
		UrgentCheckMode(sourcePtr)
		wg.Done()
	} else {
		for {
			if IsMonitorSmooth(moitorType,
				profileRequestChan,
				profileReceiveChan,
				profileChangeChan) {
				continue
			} else {
				log.Print("Request profile timeout")
				time.Sleep(5 * time.Second)
			}
		}
	}
	log.Println("Worker done. Good bye!")
}

func RegisterMasterConnect(uri string, opts *socketio_client.Options, exitMasterConChan chan bool, moitorType int, profileReceiveChan chan string, profileChangeChan chan string, profileRequestChan chan string) (isOK bool) {
	client, err := socketio_client.NewClient(uri, opts)
	if err != nil {
		log.Printf("New Client error:%v\n", err)
		return false
	}

	client.On(socket_event.Loi, func() {
		log.Printf("on error\n")
		exitMasterConChan <- true
	})
	client.On(socket_event.KetNoi, func(msg string) {
		log.Printf("Connected whith message: %v\n", msg)
		// register room
		moitorTypeString := fmt.Sprintf("%d", moitorType)
		client.Emit("register-monitor-type", moitorTypeString)
	})

	client.On(socket_event.NgatKetNoi, func() {
		log.Printf("Disconnect from server")
		exitMasterConChan <- true
	})
	client.On("profile-monitor-response", func(msg string) {
		log.Println(msg)
		profileReceiveChan <- msg
	})

	for {
		select {
		case <-exitMasterConChan:
			log.Println("Interrupt master connect, close connect..")
			return
		case profileChangeMsg := <-profileChangeChan:
			client.Emit("monitor-response", profileChangeMsg)
		case profileRequestMsg := <-profileRequestChan:
			client.Emit("profile-monitor-request", profileRequestMsg)
		default:
			time.Sleep(1 * time.Second)
		}
	}
	return false
}

func IsMonitorSmooth(monitorType int, requestChan chan string, receiveChan chan string, changeChan chan string) (isSmooth bool) {
	log.Println("Start monitor")
	const maxRetry = 3
	var retry int
	monitorTypeStr := fmt.Sprintf("%d", monitorType)
	requestChan <- monitorTypeStr
	time.Sleep(2 * time.Second)
	for  {
		select {
		case message := <- receiveChan:
			log.Println("Monitor Received", message)
			var monitorProfileData model.MonitorInputForAgent
			err := monitorProfileData.LoadFromJsonString(message)
			if err != nil {
				return false
			}
			log.Println(monitorProfileData)
			for _, profile := range monitorProfileData.ProfileList {
				log.Println("Do monitor")
				multicatsStream := fmt.Sprintf("%s:1234", profile.MulticastIP)
				_, checkcode := self_utils.CheckSourceMulticast(multicatsStream)
				if checkcode != profile.Status {
					//time.Sleep(5 * time.Second)
					//recheck
					//_, checkcode = self_utils.CheckSourceMulticast(multicatsStream)
					//if checkcode != profile.Status {
					//log.Println("Wait for recheck")
					msg := fmt.Sprintf("Status has change from %d to %d\n", profile.Status, checkcode)
					log.Println(msg)
					var signalStatus bool
					var audioStatus bool
					var videoStatus bool
					switch checkcode {
					case 0:
						videoStatus = false
						audioStatus = false
						signalStatus = false
					}
					if checkcode == 1 {
						signalStatus = true
					}
					changeStatusMessageData := model.ProfileChangeStatus{
						MonitorType:     monitorType,
						MonitorID:       profile.MonitorId,
						ProfileId:       profile.ProfileId,
						AgentId:         profile.AgentId,
						OldStatus:       profile.Status,
						NewStatus:       checkcode,
						OldSignalStatus: false,
						NewSignalStatus: signalStatus,
						OldVideoStatus:  false,
						NewVideoStatus:  videoStatus,
						OldAudioStatus:  false,
						NewAudioStatus:  audioStatus,
						EventTime:       time.Now().Unix(),
					}
					changeStatusMessageSting, _ := changeStatusMessageData.GetJsonString()
					changeChan <- changeStatusMessageSting
					//}
				}
				//log.Printf("Result status: %d", checkcode)
			}
			return true
		default:
			fmt.Println("Nothing available")
			retry += 1
			if retry > maxRetry {
				log.Printf("retry %d", retry)
				log.Println("Max retry...")
				return false
			} else {
				time.Sleep(3*time.Second)
			}
		}
	}
	time.Sleep(7*time.Second)
	return false
}

func UrgentCheckMode(sourcePtr *string) {
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
}

func GetRunmode(modePtr *string) string {
	var runMode string
	if *modePtr != "" {
		runMode = *modePtr
		if *modePtr != "daemon" && *modePtr != "urgent" {
			log.Fatalf("Run mode: daemon(default)/urgent. Not support your mode %s", *modePtr)
		}
	}
	return runMode
}

func GetWorkerCurrency(workerNumPtr *string) int {
	var threadNum int
	if *workerNumPtr != "" {
		threadNum, err := strconv.Atoi(*workerNumPtr)
		if err != nil {
			log.Fatal(err)
		} else {
			log.Printf("Worker run with %d thread(s)\n", threadNum)
		}
	}
	return threadNum
}

func GetMonitorTypeAndRegisterPidFile(monitorTypePtr *string) (int, string) {
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
	return moitorType, logFilePath
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


