package server

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func handleWS(s *Server, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade err: %v", err)
		return
	}

	// 设置为“唯一连接”，并踢掉旧的
	s.setConn(conn)
	defer func() {
		s.clearConn(conn)
		_ = conn.Close()
	}()

	for {
		mt, data, err := conn.ReadMessage()
		if err != nil {
			log.Printf("ws read err: %v", err)
			return
		}
		if mt != websocket.BinaryMessage {
			continue
		}
		dispatchMessage(s.hid, data)
	}
}
