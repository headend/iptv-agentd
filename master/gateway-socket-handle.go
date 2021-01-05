package master

import (
	socket_event "github.com/headend/share-module/configuration/socket-event"
	static_config "github.com/headend/share-module/configuration/static-config"
	agentModel "github.com/headend/share-module/model/agentd"
	socketio_client "github.com/zhouhui8915/go-socket.io-client"
	"log"
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
	client.On("profile-monitor-response", func(msg string) {
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
	for {
		select {
		case <-exitChan:
			println("break message")
			return true
		case requestMsg := <-requestChan:
			client.Emit("profile-monitor-request", requestMsg)
		case changeMsg := <-changeChan:
			client.Emit("monitor-response", changeMsg)
		default:
			//println("say hello")
			time.Sleep(1 * time.Second)
		}
	}
	return false
}
