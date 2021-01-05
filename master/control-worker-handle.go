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
				switch ctlRequestData.ControlType {
				case static_config.StartMonitorSignal:
					log.Println("[Agentd] Recieve run signal worker")
					// check process running
					pidPathFile := fmt.Sprintf("%s/run/signal.pid", static_config.InstallationPath)
					if self_utils.IsWorkerRunning(pidPathFile) {
						log.Println("Worker already run")
						return
					}
					log.Println("[Agentd] start signal worker")
					appToRUn := fmt.Sprintf("%s/iptv-agentd", static_config.BinaryPath)
					err, exitCode, stdout, stderr := shellout.RunExternalCmd(appToRUn, []string{"-m", "daemon", "-t", "signal", "-n", runThreadString}, 0)
					log.Printf("err: %s", err.Error())
					log.Printf("exitCode: %d", exitCode)
					log.Printf("stdout: %s", stdout)
					log.Printf("stderr: %s", stderr)
				case static_config.StartMonitorVideo:
					log.Println("[Agentd] run signal worker")
					pidPathFile := fmt.Sprintf("%s/run/video.pid", static_config.InstallationPath)
					if self_utils.IsWorkerRunning(pidPathFile) {
						return
					}
					appToRUn := fmt.Sprintf("%s/iptv-agentd", static_config.BinaryPath)
					err, exitCode, stdout, stderr := shellout.RunExternalCmd(appToRUn, []string{"-m", "daemon", "-t", "video", "-n", runThreadString}, 0)
					log.Printf("err: %s", err.Error())
					log.Printf("exitCode: %d", exitCode)
					log.Printf("stdout: %s", stdout)
					log.Printf("stderr: %s", stderr)
				case static_config.StartMonitorAudio:
					log.Println("[Agentd] run signal worker")
					pidPathFile := fmt.Sprintf("%s/run/audio.pid", static_config.InstallationPath)
					if self_utils.IsWorkerRunning(pidPathFile) {
						return
					}
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
					pidString, _ := pidFile.Read()
					pid, _ := strconv.Atoi(pidString)
					err := syscall.Kill(pid, 15)
					if err != nil {
						log.Println("Success stop signal monitor")
					}
				case static_config.StopMonitorVideo:
					pidFilePath := fmt.Sprintf("%s/run/video.pid", static_config.InstallationPath)
					var pidFile file_and_directory.MyFile
					pidFile.Path = pidFilePath
					pidString, _ := pidFile.Read()
					pid, _ := strconv.Atoi(pidString)
					err := syscall.Kill(pid, 15)
					if err != nil {
						log.Println("Success stop video monitor")
					}
				case static_config.StopMonitorAudio:
					pidFilePath := fmt.Sprintf("%s/run/audio.pid", static_config.InstallationPath)
					var pidFile file_and_directory.MyFile
					pidFile.Path = pidFilePath
					pidString, _ := pidFile.Read()
					pid, _ := strconv.Atoi(pidString)
					err := syscall.Kill(pid, 15)
					if err != nil {
						log.Println("Success stop audio monitor")
					}
				default:
					log.Println("Not support")
				}
			}()
		default:
			time.Sleep(1 * time.Second)
		}
	}
}
