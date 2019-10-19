package health

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"
)

type ServiceDiscovery struct {
	client       *http.Client
	paths        []PathAndResult
	ticker       *time.Ticker
	pathToConfig string
	EndChan      chan struct{}
	log          *logrus.Logger
	pattern      string
	before       string
	after        string
}

func NewServiceDiscovery(config *Config, logger *logrus.Logger) *ServiceDiscovery {
	return &ServiceDiscovery{
		client: &http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:          100,
				IdleConnTimeout:       30 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
		paths:        config.Services,
		ticker:       time.NewTicker(config.Ticker.Duration),
		pathToConfig: config.NginxConfigPath,
		log:          logger,
		EndChan:      make(chan struct{}),
		pattern:      config.PatternAddr,
		before:       config.Before,
		after:        config.After,
	}
}

func (sd *ServiceDiscovery) Run() {
	ips := make([]string, 0, len(sd.paths))
	select {
	case <-sd.ticker.C:
		for _, path := range sd.paths {
			req, err := http.NewRequest(http.MethodGet, path.Addr, nil)
			if err != nil {
				sd.log.WithField("path", path).WithError(err).Error("can't create request")

				continue
			}

			res, err := sd.client.Do(req)
			if err != nil {
				sd.log.WithField("path", path).WithError(err).Error("can't do request")

				continue
			}

			if res.StatusCode == path.ExpectedStatus {
				ips = append(ips, path.IP)
			}
			_ = res.Body.Close()
		}

		if len(ips) != 0 {
			sd.UpdateNginxConfig(ips)
		}

	case <-sd.EndChan:
		sd.log.Warning("got signal on end chan, returning...")

		return
	}
}

func (sd *ServiceDiscovery) UpdateNginxConfig(ips []string) {
	body := ""
	for _, ip := range ips {
		body += fmt.Sprintf(sd.pattern, ip)
	}

	res := sd.before + body + sd.after

	err := ioutil.WriteFile(sd.pathToConfig, []byte(res), os.ModePerm)
	if err != nil {
		sd.log.WithError(err).Error("can't update config")

		return
	}
}
