package HttpServer

import (
	"fmt"
	"log"
	"net/http"
	"simple_blockchain/pkg/handler"
)

type HttpServer struct {
	Port   string
	Server *http.Server
}

func NewHttpServer(port string, handler *handler.Handler) *HttpServer {
	router := newRouter(handler)

	var srv = &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: router.CoreRouter,
	}

	return &HttpServer{
		Port: port, Server: srv,
	}
}

func (ws *HttpServer) Run() error {
	log.Println("project started successfully")
	return ws.Server.ListenAndServe()
}

func (ws *HttpServer) Close() error {
	return ws.Server.Close()
}
