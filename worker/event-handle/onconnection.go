package event_handle

import (
	"fmt"
	socketio "github.com/googollee/go-socket.io"
)

func ListenConnection(s socketio.Conn, server *socketio.Server) error {
	s.SetContext("")
	fmt.Println("[Master] connected:", s.ID())
	fmt.Println("[Master] Allow connect from ip: ", s.RemoteAddr())
	s.Emit("connection", "Connect master successful!")
	server.JoinRoom("/", "agent", s)
	return nil
}

