package event_handle

import socketio "github.com/googollee/go-socket.io"

func OnLog(s socketio.Conn, msg string) string {
	s.SetContext(msg)
	return "recv " + msg
}

