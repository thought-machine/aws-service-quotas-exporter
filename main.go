package main

import (
	"fmt"
	"net/http"

	"github.com/jessevdk/go-flags"
	logging "github.com/sirupsen/logrus"
)

var log = logging.WithFields(logging.Fields{})

var opts struct {
	Port int `long:"port" short:"p" default:"9090" help:"Port on which to serve."`
}

func main() {
	flags.Parse(&opts)

	log.Infof("Serving on port: %d", opts.Port)
	log.Infof("Serving Prometheus metrics on /metrics")
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", opts.Port), nil))
}
