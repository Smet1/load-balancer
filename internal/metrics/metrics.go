package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sync"
)

const (
	TagTime = "response time"
)

type Statistic struct {
	ResponseTime *prometheus.CounterVec
	Requests     prometheus.Counter
}

var (
	Stats = Statistic{}
	once  = &sync.Once{}
)

func InitMetrics() {
	once.Do(func() {
		Stats.ResponseTime = prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "response_time",
		}, []string{"tag"})

		Stats.Requests = prometheus.NewCounter(prometheus.CounterOpts{
			Name: "requests",
			Help: "count of requests",
		})

		prometheus.MustRegister(Stats.ResponseTime, Stats.Requests)
	})
}
