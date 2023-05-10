package server

import (
	"net/http"
	"notifyGo/internal/api/router"
)

func NewServer(addr string) *http.Server {
	pusher := router.NewMsgPusher()
	return &http.Server{
		Addr:    addr,
		Handler: pusher.GetRouter(),
	}
}
