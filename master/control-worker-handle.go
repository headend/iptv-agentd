package master

import (
	"encoding/json"
	"fmt"
	self_utils "github.com/headend/iptv-agentd/utils"
	static_config "github.com/headend/share-module/configuration/static-config"
	file_and_directory "github.com/headend/share-module/file-and-directory"
	"github.com/headend/share-module/model"
	"github.com/headend/share-module/shellout"
	"log"
	"strconv"
	"syscall"
	"time"
)

func ControlWorkerHandle(exitControlChan chan bool, controlChan chan string) {
	for {
		select {
		case <-exitControlChan:
			return
		case ctlMsg := <-controlChan:
			//log.Printf("Received control message: %s", ctlMsg)
			var ctlRequestData *model.AgentCTLQueueRequest
			json.Unmarshal([]byte(ctlMsg), &ctlRequestData)
			go func() {
				var runThread int
				if ctlRequestData.RunThread != 0 {
					runThread = ctlRequestData.RunThread
				} else {
					runThread = 1
				}
				runThreadString := fmt.Sprintf("%d", runThread)
				appToRUn := static_config.AgentdWorkerPath
				switch ctlRequestData.ControlType {
				case static_config.UpdateWorker:
					StopSignal()
					StopVideo()
					StopAudio()
				case static_config.StartMonitorSignal:
					log.Println("[Agentd] Recieve run signal worker")
					if StartSignal(appToRUn, runThreadString) {
						return
					}
				case static_config.StartMonitorVideo:
					log.Println("[Agentd] run signal worker")
					if StartVideo(appToRUn, runThreadString) {
						return
					}
				case static_config.StartMonitorAudio:
					log.Println("[Agentd] run signal worker")
					if StartAudio(appToRUn, runThreadString) {
						return
					}
				case static_config.StopMonitorSignal:
					StopSignal()
				case static_config.StopMonitorVideo:
					StopVideo()
				case static_config.StopMonitorAudio:
					StopAudio()
				default:
					log.Printf("Not support control type: %d", ctlRequestData.ControlType)
				}
			}()
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

func StopAudio() {
	pidFilePath := static_config.WorkerAudioPid
	var pidFile file_and_directory.MyFile
	pidFile.Path = pidFilePath
	pidString, _ := pidFile.Read()
	pid, _ := strconv.Atoi(pidString)
	err := syscall.Kill(pid, 15)
	if err != nil {
		log.Println("Success stop audio monitor")
	}
}

func StopSignal() {
	pidFilePath := static_config.WorkerSignalPid
	var pidFile file_and_directory.MyFile
	pidFile.Path = pidFilePath
	pidString, _ := pidFile.Read()
	pid, _ := strconv.Atoi(pidString)
	err := syscall.Kill(pid, 15)
	if err != nil {
		log.Println("Success stop signal monitor")
	}
}

func StopVideo() {
	pidFilePath := static_config.WorkerVideoPid
	var pidFile file_and_directory.MyFile
	pidFile.Path = pidFilePath
	pidString, _ := pidFile.Read()
	pid, _ := strconv.Atoi(pidString)
	err := syscall.Kill(pid, 15)
	if err != nil {
		log.Println("Success stop video monitor")
	}
}

func StartAudio(appToRUn string, runThreadString string) bool {
	pidPathFile := static_config.WorkerAudioPid
	if self_utils.IsWorkerRunning(pidPathFile) {
		return true
	}
	err, exitCode, stdout, stderr := shellout.RunExternalCmd(appToRUn, []string{"-m", "daemon", "-t", "audio", "-n", runThreadString}, 0)
	log.Printf("err: %s", err.Error())
	log.Printf("exitCode: %d", exitCode)
	log.Printf("stdout: %s", stdout)
	log.Printf("stderr: %s", stderr)
	return false
}

func StartVideo(appToRUn string, runThreadString string) bool {
	pidPathFile := static_config.WorkerVideoPid
	if self_utils.IsWorkerRunning(pidPathFile) {
		return true
	}
	err, exitCode, stdout, stderr := shellout.RunExternalCmd(appToRUn, []string{"-m", "daemon", "-t", "video", "-n", runThreadString}, 0)
	log.Printf("err: %s", err.Error())
	log.Printf("exitCode: %d", exitCode)
	log.Printf("stdout: %s", stdout)
	log.Printf("stderr: %s", stderr)
	return false
}

func StartSignal(appToRUn string, runThreadString string) bool {
	// check process running
	pidPathFile := static_config.WorkerSignalPid
	if self_utils.IsWorkerRunning(pidPathFile) {
		log.Println("Worker already run")
		return true
	}
	log.Println("[Agentd] start signal worker")
	err, exitCode, stdout, stderr := shellout.RunExternalCmd(appToRUn, []string{"-m", "daemon", "-t", "signal", "-n", runThreadString}, 0)
	log.Printf("err: %s", err.Error())
	log.Printf("exitCode: %d", exitCode)
	log.Printf("stdout: %s", stdout)
	log.Printf("stderr: %s", stderr)
	return false
}
