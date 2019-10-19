package main

import (
	"context"
	"encoding/json"
	"flag"
	"github.com/Smet1/load-balancer/internal/metrics"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func main() {
	port := flag.String("port", ":80", "")
	portMetrics := flag.String("metrics", ":8080", "")
	flag.Parse()

	metrics.InitMetrics()

	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})

	mux := chi.NewRouter()

	mux.Use(middleware.NoCache)
	mux.Use(metrics.GetMetricsMiddleware(log))

	returnOK := true

	mux.Route("/api", func(r chi.Router) {
		r.Post("/turn", func(res http.ResponseWriter, req *http.Request) {
			returnOK = !returnOK

			res.WriteHeader(http.StatusOK)
			return
		})

		r.Get("/status", func(res http.ResponseWriter, req *http.Request) {
			if returnOK {
				res.WriteHeader(http.StatusOK)
				return
			}

			res.WriteHeader(http.StatusInternalServerError)
			return
		})

		r.Get("/dummy", func(res http.ResponseWriter, req *http.Request) {
			sleep := 10 * time.Second
			if returnOK {
				sleep = 1 * time.Second
			}

			time.Sleep(sleep)

			answer := struct {
				ID     string        `json:"id"`
				Waited time.Duration `json:"waited"`
			}{
				ID:     uuid.New().String(),
				Waited: sleep,
			}

			b, err := json.Marshal(answer)
			if err != nil {
				log.WithError(err).Error("can't marshal anser")

				res.WriteHeader(http.StatusUnprocessableEntity)
				return
			}

			_, err = res.Write(b)
			if err != nil {
				log.WithError(err).Error("can't marshal anser")

				res.WriteHeader(http.StatusUnprocessableEntity)
				return
			}
		})
	})

	server := http.Server{
		Handler: mux,
		Addr:    *port,
	}

	go func() {
		log.Infof("dummy slow service started on port %s", *port)
		if err := server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				log.Info("graceful shutdown")
			} else {
				log.WithError(err).Fatal("sync service")
			}
		}
	}()

	go func() {
		metricsMux := chi.NewRouter()
		metricsMux.Handle("/metrics", promhttp.Handler())

		log.Infof("metrics for dummy slow service started on port %s", *portMetrics)
		if err := http.ListenAndServe(*portMetrics, metricsMux); err != nil {
			if err == http.ErrServerClosed {
				log.Info("graceful shutdown")
			} else {
				log.WithError(err).Fatal("metrics")
			}
		}
	}()

	sgnl := make(chan os.Signal, 1)
	signal.Notify(sgnl,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	stop := <-sgnl

	if err := server.Shutdown(context.Background()); err != nil {
		log.WithError(err).Error("error on shutdown")
	}

	log.WithField("signal", stop).Info("stopping")
}
