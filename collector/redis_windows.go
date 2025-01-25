//go:build windows
// +build windows

package collector

import (
	"github.com/orange-cloudfoundry/custom_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type CollectorRedis struct {
	metricsConfig config.MetricsItem
}

func NewCollectorRedis(config config.MetricsItem) *CollectorRedis {
	return &CollectorRedis{
		metricsConfig: config,
	}
}

func NewPrometheusRedisCollector(config config.MetricsItem) (prometheus.Collector, error) {
	log.Warnf("Redis collector is not supported on Windows")
	return nil, nil
}

func (e CollectorRedis) Config() config.MetricsItem {
	return e.metricsConfig
}

func (e CollectorRedis) Name() string {
	return "redis"
}

func (e CollectorRedis) Desc() string {
	return "Redis collector (disabled on Windows)"
}

func (e CollectorRedis) Run(ch chan<- prometheus.Metric) error {
	return nil
}