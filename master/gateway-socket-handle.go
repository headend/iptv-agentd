package master

import (
	"fmt"
	socket_event "github.com/headend/share-module/configuration/socket-event"
	static_config "github.com/headend/share-module/configuration/static-config"
	curl_http "github.com/headend/share-module/curl-http"
	"github.com/headend/share-module/model"
	agentModel "github.com/headend/share-module/model/agentd"
	"github.com/headend/share-module/shellout"
	socketio_client "github.com/zhouhui8915/go-socket.io-client"
	"log"
	"net/url"
	"os"
	"time"
)

func RegisterGatewayClientSocket(uri string,
	opts *socketio_client.Options,
	exitChan chan bool,
	requestChan chan string,
	receiveChan chan string,
	receiveSignalChan chan string,
	receiveVideoChan chan string,
	receiveAudioChan chan string,
	changeChan chan string,
	controlChan chan string) (isRetry bool) {
	client, err := socketio_client.NewClient(uri, opts)
	if err != nil {
		log.Printf("[Agentd] NewClient error:%v\n", err)
		return true
	}

	client.On(socket_event.Loi, func() {
		log.Printf("[Agentd] on error\n")
		// enable exist mode
		exitChan <- true
	})
	client.On(socket_event.KetNoi, func(msg string) {
		log.Printf("[Agentd] Connected whith message: %v\n", msg)

	})

	client.On(socket_event.NgatKetNoi, func() {
		log.Printf("[Agentd] Disconnect from server")
		exitChan <- true
	})

	client.On(socket_event.DieuKhien, func(msg string) {
		//log.Println(msg)
		controlChan <- msg
	})
	client.On(socket_event.ProfileResponse, func(msg string) {
		// Transfer message to worker
		var data agentModel.MonitorInputForAgent
		data.LoadFromJsonString(msg)
		switch data.MonitorType {
		case static_config.Video:
			receiveVideoChan <- msg
		case static_config.Audio:
			receiveAudioChan <- msg
		case static_config.Signal:
			receiveSignalChan <- msg
		default:
			receiveChan <- msg
		}

	})

	// ping
	client.On(socket_event.PingPing, func(msg string) {
		log.Println("ping")
		// Transfer message to worker
		appToRUn := fmt.Sprintf("%s/iptv-agentd-worker", static_config.BinaryPath)
		err, exitCode, stdout, stderr := shellout.RunExternalCmd(appToRUn, []string{"-v", "version"}, 5)
		if err != nil {
			log.Println(err)
		} else {
			if exitCode == 0 {
				client.Emit(socket_event.PongPong, stdout)
			} else {
				log.Printf("%s - %s", stdout, stderr)
			}
		}
	})

	client.On(socket_event.UpdateWorker, func(msg string) {
		// Transfer message to worker
		log.Println("Recieve update....")
		var updateDataRequest  model.WorkerUpdateSignal
		err := updateDataRequest.LoadFromJsonString(msg)
		if err != nil {
			log.Println(err)
		} else {
			// download file
			u, _ := url.Parse(uri)

			url := fmt.Sprintf("%s://%s/%s", u.Scheme, u.Host, updateDataRequest.FileName)
			log.Printf("Download new version to %s", updateDataRequest.FilePath + "tmp")
			err2:=curl_http.DownloadFile(url, updateDataRequest.FilePath + "tmp")
			if err2 != nil {
				log.Println(err2)
				log.Println("Resync worker...")
				client.Emit("sync-worker", "sync worker")
			} else {
				_ = os.Rename(updateDataRequest.FilePath + "tmp", updateDataRequest.FilePath)
				os.Chmod(updateDataRequest.FilePath, 0755)
				//check file size
				// check md5
				// start worker
				log.Println("Resync worker...")
				client.Emit("sync-worker", "sync worker")
			}

		}
	})

	for {
		select {
		case <-exitChan:
			println("break message")
			return true
		case requestMsg := <-requestChan:
			client.Emit(socket_event.ProfileRequest, requestMsg)
		case changeMsg := <-changeChan:
			client.Emit(socket_event.MonitorResponse, changeMsg)
		default:
			//println("say hello")
			time.Sleep(1 * time.Second)
		}
	}
	return false
}


