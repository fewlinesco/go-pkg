package web

import (
	_ "expvar" // Register the expvar handlers
	"net/http"
	_ "net/http/pprof" // Register the pprof handlers
	"time"
)

type ServerConfig struct {
	Address         string `json:"address"`
	ReadTimeout     int    `json:"read_timeout"`
	WriteTimeout    int    `json:"write_timeout"`
	ShutdownTimeout int    `json:"shutdown_timeout"`
}

var DefaultServerConfig = ServerConfig{
	Address:         ":8080",
	ReadTimeout:     30,
	WriteTimeout:    30,
	ShutdownTimeout: 45,
}

var DefaultMonitoringConfig = ServerConfig{
	Address:         ":8081",
	ReadTimeout:     30,
	WriteTimeout:    30,
	ShutdownTimeout: 45,
}

func NewServer(config ServerConfig, router http.Handler) *http.Server {
	server := http.Server{
		Addr:         config.Address,
		Handler:      router,
		ReadTimeout:  time.Duration(config.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(config.WriteTimeout) * time.Second,
	}

	return &server
}

func NewMonitoringServer(config ServerConfig) *http.Server {
	return NewServer(config, http.DefaultServeMux)
}