package main

import (
	"fmt"
	"github.com/aibotsoft/micro/config"
	"github.com/aibotsoft/micro/config_client"
	"github.com/aibotsoft/micro/logger"
	"github.com/aibotsoft/micro/sqlserver"
	"github.com/aibotsoft/sportmarket-service/pkg/store"
	"github.com/aibotsoft/sportmarket-service/services/auth"
	"github.com/aibotsoft/sportmarket-service/services/handler"
	"github.com/aibotsoft/sportmarket-service/services/server"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.New()
	log := logger.New()
	log.Infow("Begin service", "config", cfg.Service)
	db := sqlserver.MustConnectX(cfg)
	sto := store.New(cfg, log, db)
	conf := config_client.New(cfg, log)
	au := auth.New(cfg, log, sto, conf)
	//go au.AuthJob()
	h := handler.New(cfg, log, sto, au)
	go h.BalanceJob()
	go h.WebsocketJob()
	go h.BetListJob()

	s := server.New(cfg, log, h)
	// Инициализируем Close
	errc := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	//c := collector.New(cfg, log, sto, au)
	//go c.CollectJob()

	go func() { errc <- s.Serve() }()
	defer func() { s.Close() }()
	log.Info("exit: ", <-errc)
}
