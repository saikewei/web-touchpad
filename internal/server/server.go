package server

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"sync"
	"web-touchpad/internal/hid"

	"github.com/gorilla/websocket"
)

//go:embed web/*
var webFS embed.FS

type Server struct {
	hid  *hid.Device
	http *http.Server

	mu   sync.Mutex
	conn *websocket.Conn
}

func New(dev *hid.Device) *Server {
	return &Server{hid: dev}
}

func (s *Server) Run(addr string) error {
	sub, err := fs.Sub(webFS, "web")
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(sub)))
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWS(s, w, r)
	})

	s.http = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	log.Printf("Listening on %s", addr)
	return s.http.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	// 先关闭 WS
	s.closeConn()

	// 再关闭 HTTP
	if s.http == nil {
		return nil
	}
	return s.http.Shutdown(ctx)
}

func (s *Server) setConn(c *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 踢掉旧连接
	if s.conn != nil {
		_ = s.conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "replaced by new connection"),
		)
		_ = s.conn.Close()
	}
	s.conn = c
}

func (s *Server) clearConn(c *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn == c {
		s.conn = nil
	}
}

func (s *Server) closeConn() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn != nil {
		_ = s.conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "server shutdown"),
		)
		_ = s.conn.Close()
		s.conn = nil
	}
}
