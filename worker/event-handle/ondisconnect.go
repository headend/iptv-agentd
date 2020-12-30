package event_handle

import (
	"fmt"
	socketio "github.com/googollee/go-socket.io"
)

func OnDisconnection(s socketio.Conn, reason string) {
	fmt.Println("[Worker] Disconnect from: ", s.RemoteAddr())
	fmt.Println("[Worker] closed", reason)
}
