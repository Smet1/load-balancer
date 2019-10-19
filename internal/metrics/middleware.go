package metrics

import (
	"github.com/sirupsen/logrus"
	"go.opencensus.io/plugin/ochttp"
	"net/http"
	"time"
)

func GetMetricsMiddleware(log *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return ochttp.Handler{
			Handler: http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
				begin := time.Now()
				Stats.Requests.Inc()

				next.ServeHTTP(res, req)

				Stats.ResponseTime.WithLabelValues(TagTime).Add(time.Since(begin).Seconds())
			}),
		}.Handler
	}
}
