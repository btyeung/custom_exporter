//go:build windows
// +build windows

package collector

import (
	"github.com/orange-cloudfoundry/custom_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const (
	CollectorBashName = "bash"
	CollectorBashDesc = "Metrics from shell collector in the custom exporter."
)

type CollectorBash struct {
	metricsConfig config.MetricsItem
}

func NewCollectorBash(config config.MetricsItem) *CollectorBash {
	return &CollectorBash{
		metricsConfig: config,
	}
}

func NewPrometheusBashCollector(config config.MetricsItem) (prometheus.Collector, error) {
	myCol := NewCollectorHelper(
		NewCollectorBash(config),
	)

	log.Infof("Collector Added: Type '%s' / Name '%s' / Credentials '%s'", CollectorBashName, config.Name, config.Credential.Name)

	return myCol, myCol.Check(nil)
}

func (e CollectorBash) Config() config.MetricsItem {
	return e.metricsConfig
}

func (e CollectorBash) Name() string {
	return CollectorBashName
}

func (e CollectorBash) Desc() string {
	return CollectorBashDesc
}

func (e CollectorBash) Run(ch chan<- prometheus.Metric) error {
	var output []byte
	var err error
	var command string
	var args []string
	var cmd *exec.Cmd

	os.Setenv("CREDENTIALS_NAME", e.metricsConfig.Credential.Name)
	os.Setenv("CREDENTIALS_COLLECTOR", e.metricsConfig.Credential.Collector)
	os.Setenv("CREDENTIALS_DSN", e.metricsConfig.Credential.Dsn)
	os.Setenv("CREDENTIALS_PATH", e.metricsConfig.Credential.Path)
	os.Setenv("CREDENTIALS_URI", e.metricsConfig.Credential.Uri)

	if e.metricsConfig.Credential.User != "" {
		log.Warnf("User credentials are not supported on Windows. Ignoring user: %s", e.metricsConfig.Credential.User)
	}

	regexCmd := regexp.MustCompile("'.+'|\".+\"|\\S+")

	for _, c := range e.metricsConfig.Commands {
		args = regexCmd.FindAllString(c, -1)
		command, args = args[0], args[1:]

		log.Debugf("Parsed command : %s -- %v", command, args)
		log.Debugf("Checking command/script exists : \"%s\"...", command)

		_, err = exec.LookPath(command)
		if err != nil {
			log.Errorf("Error with metric \"%s\" while checking command exists \"%s\" : %s", e.metricsConfig.Name, c, err.Error())
			return err
		}

		log.Debugf("Running command \"%s\" with params \"%s\"...", command, args)

		cmd = exec.Command("cmd", "/C", command)
		cmd.Args = append(cmd.Args, args...)
		cmd.Env = os.Environ()
		cmd.Stdin = strings.NewReader(string(output))

		output, err = cmd.CombinedOutput()

		if err != nil {
			log.Errorf("Error with metric \"%s\" while running command \"%s\" : %v : %v", e.metricsConfig.Name, c, err, string(output))
			return err
		}

		log.Debugf("Result command \"%s\" : \"%s\"", command, string(output))
	}

	log.Debugf("Run metric \"%s\" command '%s'", e.metricsConfig.Name, command)
	log.Debugln("Result:", "\n"+string(output))

	return e.parse(ch, string(output))
}

func (e CollectorBash) parse(ch chan<- prometheus.Metric, output string) error {
	var err error

	err = nil
	sep := e.metricsConfig.Separator
	nb := len(e.metricsConfig.Mapping) + 1

	for _, l := range strings.Split(output, "\n") {
		if len(strings.TrimSpace(l)) < nb {
			continue
		}

		log.Debugf("Parsing line: \"%s\"...", l)

		l = strings.Trim(strings.TrimSpace(l), sep)

		if errline := e.parseLine(ch, strings.Split(l, sep)); errline != nil {
			log.Errorf("Error with metric \"%s\" while parsing line : %s", e.metricsConfig.Name, errline.Error())
			err = errline
		}
	}

	return err
}

func (e *CollectorBash) parseLine(ch chan<- prometheus.Metric, fields []string) error {
	var (
		mapping   []string
		labelVal  []string
		metricVal float64
		err       error
	)

	mapping = e.metricsConfig.Mapping
	labelVal = make([]string, len(mapping))
	err = nil

	for i, value := range fields {
		value = strings.TrimSpace(value)

		if (i + 1) > len(mapping) {
			if metricVal, err = strconv.ParseFloat(value, 64); err != nil {
				metricVal = float64(0)
			}
		} else {
			labelVal[i] = value
		}
	}

	if err != nil {
		log.Debugf("Return error : '%s'", err.Error())
		return err
	}

	prom_desc := PromDesc(e)
	log.Debugf("Add Metric \"%s\" : Tag '%s' / TagValue '%s' / Value '%v'", prom_desc, mapping, labelVal, metricVal)

	metric := prometheus.MustNewConstMetric(
		prometheus.NewDesc(prom_desc, e.metricsConfig.Name, mapping, nil),
		e.metricsConfig.Value_type, metricVal, labelVal...,
	)

	select {
	case ch <- metric:
		log.Debug("Return no error...")
		return nil
	default:
		log.Info("Cannot write to channel...")
	}

	return err
}