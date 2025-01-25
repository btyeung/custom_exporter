//go:build windows
// +build windows

package collector

import (
	"github.com/orange-cloudfoundry/custom_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type CollectorMysql struct {
	metricsConfig config.MetricsItem
}

func NewCollectorMysql(config config.MetricsItem) *CollectorMysql {
	return &CollectorMysql{
		metricsConfig: config,
	}
}

func NewPrometheusMysqlCollector(config config.MetricsItem) (prometheus.Collector, error) {
	log.Warnf("MySQL collector is not supported on Windows")
	return nil, nil
}

func (e CollectorMysql) Config() config.MetricsItem {
	return e.metricsConfig
}

func (e CollectorMysql) Name() string {
	return "mysql"
}

func (e CollectorMysql) Desc() string {
	return "MySQL collector (disabled on Windows)"
}

func (e CollectorMysql) Run(ch chan<- prometheus.Metric) error {
	return nil
}