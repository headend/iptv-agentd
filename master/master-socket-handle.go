package master

import (
	"fmt"
	"github.com/googollee/go-socket.io"
	master_event_handle "github.com/headend/iptv-agentd/master/event-handle"
	self_utils "github.com/headend/iptv-agentd/utils"
	"github.com/headend/share-module/configuration"
	socket_event "github.com/headend/share-module/configuration/socket-event"
	socketio_client "github.com/zhouhui8915/go-socket.io-client"
	"github.com/headend/share-module/configuration/static-config"
	"log"
	"net/http"
	"strconv"
)


func RegisterMasterSocket(gwClient *socketio_client.Client, masterSocket *socketio.Server, conf configuration.Conf, chDestroyInternalSocker *chan bool) {
	gwHost, gwPort := self_utils.GetMasterConnectionInfo(conf)
	masterSocket.OnConnect("/", func(s socketio.Conn) error {
		return master_event_handle.ListenConnection(s, masterSocket)
	})

	masterSocket.OnEvent("/", socket_event.ThongBao, func(s socketio.Conn, msg string) {
		master_event_handle.OnNotice(s, msg)
	})

	masterSocket.OnEvent(socket_event.NhanLog, socket_event.NhanLog, func(s socketio.Conn, msg string) string {
		return master_event_handle.OnLog(s, msg)
	})
	masterSocket.OnEvent("/", socket_event.KetQuaThucThiLenh, func(s socketio.Conn, msg string) {
		content := fmt.Sprintf("On %s result: %s", s.RemoteAddr(), msg)
		log.Printf("Delivery message to gateway: %s", content)
		gwClient.Emit(socket_event.KetQuaThucThiLenh, content)
	})

	masterSocket.OnEvent("/", "bye", func(s socketio.Conn) string {
		last := s.Context().(string)
		s.Emit("bye", last)
		s.Close()
		return last
	})

	masterSocket.OnEvent("/", socket_event.DieuKhien, func(s socketio.Conn, signal int) {
		masterSocket.BroadcastToRoom("/", socket_event.NhomChung, socket_event.DieuKhien, signal)
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
		switch i {
		case static_config.Audio:
			masterSocket.JoinRoom("/", "audio", s)
		case static_config.Video:
			masterSocket.JoinRoom("/", "video", s)
		default:
			masterSocket.JoinRoom("/", "signal", s)
		}
		log.Println(masterSocket.Rooms("/"))
	})

	masterSocket.OnEvent("/", "profile-monitor-request", func(s socketio.Conn, msg string) {
		// Formard message to gateway
		log.Printf("[Master] Receive request monitor type: %s\n", msg)
		err := gwClient.Emit("profile-monitor-request", msg)
		if err != nil {
			log.Printf("[Master] critical error: %s\n", err.Error())
		}
	})

	masterSocket.OnEvent("/", "monitor-response", func(s socketio.Conn, msg string) {
		log.Println(msg)
		err := gwClient.Emit("monitor-response", msg)
		if err != nil {
			log.Printf("[Master] critical error: %s\n", err.Error())
		}
	})

	masterSocket.OnEvent("/", "monitor-response", func(s socketio.Conn, msg string) {
		err := gwClient.Emit("monitor-response", msg)
		if err != nil {
			log.Printf("[Master] critical error: %s\n", err.Error())
		}
	})

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


