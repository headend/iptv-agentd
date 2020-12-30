package master

import (
	"fmt"
	socketio "github.com/googollee/go-socket.io"
	socket_event "github.com/headend/share-module/configuration/socket-event"
	socketio_client "github.com/zhouhui8915/go-socket.io-client"
	"log"
)

func OnExecuteRequestHandle(msg string, gwClient *socketio_client.Client, masterSocket *socketio.Server) {
	go func() {
		log.Printf("From master: %s", msg)
		masterSocket.BroadcastToRoom("/", socket_event.NhomChung, socket_event.ThucThiLenh, msg)
		log.Printf("There are %d worker recieve", masterSocket.Count())
		log.Printf("Total Rom on /: %s", masterSocket.Rooms("/"))
		masterSocket.ForEach("/", socket_event.NhomChung, func(conn socketio.Conn) {
			log.Println(conn.ID())
			log.Println(conn.RemoteAddr())
		})

		log.Printf("Total client on rom %s = %d", socket_event.NhomChung, masterSocket.RoomLen("/", socket_event.NhomChung))
		msgNotice := fmt.Sprintf("Send to %d worker", masterSocket.Count())
		gwClient.Emit(msgNotice)

	}()
}
