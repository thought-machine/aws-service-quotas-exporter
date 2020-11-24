package main

import (
	"fmt"
	"net/http"

	"github.com/jessevdk/go-flags"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	logging "github.com/sirupsen/logrus"
	"github.com/thought-machine/aws-service-quotas-exporter/service_exporter/serviceexporter"
)

var log = logging.WithFields(logging.Fields{})

var opts struct {
	Port    int    `long:"port" short:"p" default:"9090" description:"Port on which to serve."`
	Region  string `long:"region" short:"r" env:"AWS_REGION" required:"true" description:"AWS region name"`
	Profile string `long:"profile" short:"f" env:"AWS_PROFILE" default:"" description:"Named AWS profile to be used"`
}

func main() {
	flags.Parse(&opts)

	quotasExporter, err := serviceexporter.NewServiceQuotasExporter(opts.Region, opts.Profile)
	if err != nil {
		log.Fatalf("Failed to create exporter: %s", err)
	}

	prometheus.Register(quotasExporter)

	log.Infof("Serving on port: %d", opts.Port)
	log.Infof("Serving Prometheus metrics on /metrics")
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", opts.Port), nil))
}
