package master

import (
	"fmt"
	socketio "github.com/googollee/go-socket.io"
	master_event_handle "github.com/headend/iptv-agentd/master/event-handle"
	self_utils "github.com/headend/iptv-agentd/utils"
	"github.com/headend/share-module/configuration"
	socket_event "github.com/headend/share-module/configuration/socket-event"
	static_config "github.com/headend/share-module/configuration/static-config"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

func RegisterMasterSocket(wg *sync.WaitGroup, conf configuration.Conf, profileRequestChan chan string, profileChangeChan chan string, profileReceiveChan chan string, profileSignalReceiveChan chan string, profileVideoReceiveChan chan string, profileAudioReceiveChan chan string) {
	exitMasterChan := make(chan bool)
	masterSocket, err := socketio.NewServer(nil)
	if err != nil {
		log.Print(err)
		wg.Done()
		wg.Done()
		return
	}
	gwHost, gwPort := self_utils.GetMasterConnectionInfo(conf)
	masterSocket.OnConnect("/", func(s socketio.Conn) error {
		return master_event_handle.ListenConnection(s, masterSocket)
	})

	masterSocket.OnEvent("/", "bye", func(s socketio.Conn) string {
		last := s.Context().(string)
		s.Emit("bye", last)
		s.Close()
		return last
	})

	masterSocket.OnError("/", func(s socketio.Conn, e error) {
		master_event_handle.OnErr(s, e)
	})

	masterSocket.OnDisconnect("/", func(s socketio.Conn, reason string) {
		master_event_handle.OnDisconnection(s, reason)
	})

	masterSocket.OnEvent("/", "register-monitor-type", func(s socketio.Conn, msg string) {
		// Formard message to gateway
		i, err := strconv.Atoi(msg)
		if err != nil {
			log.Printf("[Master] critical error: %s\n", err.Error())
		}
		var romName string
		switch i {
		case static_config.Audio:
			romName = "audio"
		case static_config.Video:
			romName = "video"
		default:
			romName = "signal"
		}
		masterSocket.JoinRoom("/", romName, s)
		log.Printf("Success to join client %s to rom %s", s.LocalAddr().String(), romName)
	})

	masterSocket.OnEvent("/", socket_event.ProfileRequest, func(s socketio.Conn, msg string) {
		log.Printf("[Master] %s\n", msg)
		// Formard message to gateway
		profileRequestChan <- msg
	})

	masterSocket.OnEvent("/", socket_event.MonitorResponse, func(s socketio.Conn, msg string) {
		log.Printf("[Master] %s\n", msg)
		profileChangeChan <- msg
	})

	go func() {
		for {
			select {
			case <-exitMasterChan:
				log.Println("[Mater] Interrupt master socket")
				return
			case profileReceiveMsg := <-profileReceiveChan:
				log.Println("[Master] send to all worker")
				masterSocket.BroadcastToRoom("/", socket_event.NhomChung, socket_event.ProfileResponse, profileReceiveMsg)
			case profileReceiveMsg := <-profileSignalReceiveChan:
				log.Println("[Master] send to signal worker")
				masterSocket.BroadcastToRoom("/", "signal", socket_event.ProfileResponse, profileReceiveMsg)
				log.Print(profileReceiveMsg)
				//log.Print(masterSocket.RoomLen("/", "signal"))
			case profileReceiveMsg := <-profileVideoReceiveChan:
				log.Println("[Master] send to video worker")
				masterSocket.BroadcastToRoom("/", "video", socket_event.ProfileResponse, profileReceiveMsg)
			case profileReceiveMsg := <-profileAudioReceiveChan:
				log.Println("[Master] send to audio worker")
				masterSocket.BroadcastToRoom("/", "audio", socket_event.ProfileResponse, profileReceiveMsg)
			default:
				time.Sleep(1 * time.Second)
			}
		}
	}()

	go masterSocket.Serve()
	defer masterSocket.Close()
	http.Handle("/socket.io/", masterSocket)
	//http.Serve(ln,server)
	http.Handle("/", http.FileServer(http.Dir("./asset")))

	listenAddress := fmt.Sprintf("%s:%d", gwHost, gwPort)
	log.Println("[Master] Serving at ", listenAddress)
	// runserver here
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}