package httpserver

import (
	"context"
	"fmt"
	"sync"
)

type ServersMessage struct {
	Server  *Server
	Message string
	State   ServersMessageState
}

type ServersMessageState int

const (
	ServersMessageStateAllShutdown ServersMessageState = iota
	ServersMessageStateStarted
	ServersMessageStateStopped
	ServersMessageStateFailed
)

type Servers struct {
	servers  []*Server
	wait     sync.WaitGroup
	messages chan ServersMessage
}

func NewServers() *Servers {
	return &Servers{
		servers:  []*Server{},
		messages: make(chan ServersMessage),
	}
}

func (s *Servers) Add(server *Server) {
	s.wait.Add(1)
	s.servers = append(s.servers, server)
}

func (s *Servers) Start() {
	for _, server := range s.servers {
		server.Start()

		s.messages <- ServersMessage{
			Server:  server,
			State:   ServersMessageStateStarted,
			Message: fmt.Sprintf("server started on port %s", server.HTTPServer.Addr),
		}

		go func(server *Server) {
			select {
			case err := <-server.Stopped():
				s.messages <- ServersMessage{Server: server, State: ServersMessageStateStopped, Message: err.Error()}
			case err := <-server.Failed():
				s.messages <- ServersMessage{Server: server, State: ServersMessageStateFailed, Message: err.Error()}
			}

			s.wait.Done()
		}(server)
	}
}

func (s *Servers) WaitForShutdown(ctx context.Context) {
	for _, server := range s.servers {
		server.Shutdown(ctx)
	}

	s.wait.Wait()
	s.messages <- ServersMessage{State: ServersMessageStateAllShutdown, Message: "all servers stopped"}
}

func (s *Servers) ListenMessages() <-chan ServersMessage {
	return s.messages
}
