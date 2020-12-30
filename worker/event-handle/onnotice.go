package event_handle

import (
	"fmt"
	socketio "github.com/googollee/go-socket.io"
)

func OnNotice(s socketio.Conn, msg string) {
	fmt.Println("notice:", msg)
	s.Emit("reply", "have "+msg)
}

