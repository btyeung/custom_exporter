package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/orange-cloudfoundry/custom_exporter/collector"
	"github.com/orange-cloudfoundry/custom_exporter/custom_config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
)

var ArgsRequire []string
var ArgsSeen map[string]bool

var showVersion = flag.Bool(
	"version",
	false,
	"Print version information.",
)

var listenAddress = flag.String(
	"web.listen-address",
	":9213",
	"Address to listen on for web interface and telemetry.",
)

var metricPath = flag.String(
	"web.telemetry-path",
	"/metrics",
	"Path under which to expose metrics.",
)

var configFile = flag.String(
	"collector.config",
	"",
	"Path to config.yml file to read custom exporter definition.",
)

func init() {
	ArgsRequire = []string{
		"collector.config",
	}

	ArgsSeen = make(map[string]bool, 0)

	prometheus.MustRegister(version.NewCollector(custom_config.Namespace + "_" + custom_config.Exporter))
}

func main() {
	fmt.Fprintln(os.Stdout, version.Info())
	fmt.Fprintln(os.Stdout, version.BuildContext())

	if *showVersion {
		os.Exit(0)
	}

	if ok := checkRequireArgs(); !ok {
		os.Exit(2)
	}

	if _, err := os.Stat(*configFile); err != nil {
		log.Errorln("Error:", err.Error())
		os.Exit(2)
	}

	var myConfig *custom_config.Config

	if cnf, err := custom_config.NewConfig(*configFile); err != nil {
		log.Fatalf("FATAL: %s", err.Error())
	} else {
		myConfig = cnf
	}

	prometheus.MustRegister(createListCollectors(myConfig)...)

	http.Handle(*metricPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html><head><title>Custom exporter</title></head><body><h1>Custom exporter</h1><p><a href='` + *metricPath + `'>Metrics</a></p></body></html>`))
	})

	log.Infoln("Listening on", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

func checkRequireArgs() bool {
	var res bool

	res = true

	flag.Parse()
	flag.Visit(func(f *flag.Flag) { ArgsSeen[f.Name] = true })

	for _, req := range ArgsRequire {
		if !ArgsSeen[req] {
			fmt.Fprintf(os.Stderr, "missing required -%s argument/flag\n", req)
			res = false
		}
	}

	if !res {
		fmt.Fprintf(os.Stdout, "")
		fmt.Fprintf(os.Stdout, "")
		flag.Usage()
	}

	return res
}

func createListCollectors(c *custom_config.Config) []prometheus.Collector {
	var result []prometheus.Collector

	for _, cnf := range c.Metrics {
		if col := createNewCollector(&cnf); col != nil {
			result = append(result, col)
		}
	}

	if len(result) < 1 {
		log.Fatalf("Error : the metrics list is empty !!")
	}

	return result
}

func createNewCollector(m *custom_config.MetricsItem) prometheus.Collector {
	var col prometheus.Collector
	var err error

	switch m.Credential.Collector {
	case "bash":
		col, err = collector.NewPrometheusBashCollector(*m)
	case "mysql":
		col, err = collector.NewPrometheusMysqlCollector(*m)
	case "redis":
		col, err = collector.NewPrometheusRedisCollector(*m)
	default:
		return nil
	}

	if err != nil {
		log.Errorf("Error:", err.Error())
		return nil
	}

	return col
}
