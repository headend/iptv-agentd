package event_handle

import (
	"fmt"
	socketio "github.com/googollee/go-socket.io"
	socket_event "github.com/headend/share-module/configuration/socket-event"
)

func ListenConnection(s socketio.Conn, server *socketio.Server) error {
	s.SetContext("")
	fmt.Println("[Master] connected:", s.ID())
	fmt.Println("[Master] Allow connect from ip: ", s.RemoteAddr())
	s.Emit("connection", "Welcome to join master!")
	server.JoinRoom("/", socket_event.NhomChung, s)
	return nil
}

