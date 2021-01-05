package event_handle

import (
	"fmt"
	socketio "github.com/googollee/go-socket.io"
)

func OnDisconnection(s socketio.Conn, reason string) {
	fmt.Println("[Master] Disconnect from: ", s.RemoteAddr())
	fmt.Println("closed", reason)
	s.Close()
}
