package main

import (
	"flag"
	"github.com/Smet1/load-balancer/internal/health"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	configPath := flag.String("config", "./config.yaml", "")
	flag.Parse()

	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})

	config := &health.Config{}
	err := health.ReadConfig(*configPath, config)
	if err != nil {
		log.WithError(err).Fatal("can't read config")
	}

	log.WithField("config", config).Info("started with")

	sd := health.NewServiceDiscovery(config, log)
	go sd.Run()

	sgnl := make(chan os.Signal, 1)
	signal.Notify(sgnl,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	stop := <-sgnl

	sd.EndChan <- struct{}{}

	log.WithField("signal", stop).Info("stopping")
}
