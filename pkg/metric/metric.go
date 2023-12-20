package metric

import (
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"strconv"
)

type Metrics interface {
	IncreaseHits(status int, method, path string)
	ObserveResponseTime(status int, method, path string, observeTime float64)
}

type PrometheusMetrics struct {
	HitsTotal prometheus.Counter
	Hits      *prometheus.CounterVec
	Times     *prometheus.HistogramVec
}

func (promMetric *PrometheusMetrics) IncreaseHits(status int, method, path string) {
	promMetric.HitsTotal.Inc()
	promMetric.Hits.WithLabelValues(strconv.Itoa(status), method, path).Inc()
}

func (promMetric *PrometheusMetrics) ObserveResponseTime(status int, method, path string, observeTime float64) {
	promMetric.Times.WithLabelValues(strconv.Itoa(status), method, path).Observe(observeTime)
}

func CreateMetrics(address string, name string) (Metrics, error) {
	var promMetric PrometheusMetrics
	promMetric.HitsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: name + "_hits_total",
	})

	if err := prometheus.Register(promMetric.HitsTotal); err != nil {
		return nil, err
	}

	promMetric.Hits = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: name + "_hits",
	}, []string{"status", "method", "path"})

	if err := prometheus.Register(promMetric.Hits); err != nil {
		return nil, err
	}

	promMetric.Times = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: name + "_times",
	}, []string{"status", "method", "path"})

	if err := prometheus.Register(promMetric.Times); err != nil {
		return nil, err
	}

	go func() {
		router := echo.New()
		router.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
		log.Printf("Metrics server is running on port: %s", address)
		if err := router.Start(address); err != nil {
			log.Fatal(err)
		}
	}()

	return &promMetric, nil
}
