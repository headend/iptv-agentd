package event_handle

import (
	"fmt"
	socketio "github.com/googollee/go-socket.io"
)

func OnErr(s socketio.Conn, e error) (int, error) {
	return fmt.Println("[Master] meet error:", e)
}

