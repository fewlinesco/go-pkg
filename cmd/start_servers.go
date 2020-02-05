package cmd

import (
	"context"
	"github.com/fewlinesco/go-pkg/httpserver"
	"github.com/fewlinesco/go-pkg/logging"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func StartServers(servers *httpserver.Servers, logger logging.Logger) int {
	go func(logger logging.Logger, servers *httpserver.Servers) {
		for {
			msg := <-servers.ListenMessages()
			switch msg.State {
			case httpserver.ServersMessageStateAllShutdown:
				logger.Info(msg.Message)
			case httpserver.ServersMessageStateStarted:
				logger.With(logging.String("server", msg.Server.Name)).Info(msg.Message)
			case httpserver.ServersMessageStateFailed:
				logger.With(logging.String("server", msg.Server.Name)).Error(msg.Message)
			case httpserver.ServersMessageStateStopped:
				logger.With(logging.String("server", msg.Server.Name)).Info(msg.Message)
			default:
				logger.Info(msg.Message)
			}
		}
	}(logger, servers)

	servers.Start()

	interruptions := make(chan os.Signal, 1)
	signal.Notify(interruptions, syscall.SIGINT, syscall.SIGTERM)

	<-interruptions

	logger.Info("wait for current connections to finish")

	serverGracePeriodCtx, abort := context.WithTimeout(context.Background(), 15*time.Second)
	defer abort()

	servers.WaitForShutdown(serverGracePeriodCtx)

	return 0
}
