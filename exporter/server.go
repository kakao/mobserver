package exporter

import (
	"net/http"
	"os"
	"time"

	"github.com/prometheus/common/promlog"
	"github.com/prometheus/exporter-toolkit/web"
	"github.com/sirupsen/logrus"
)

// ServerOpts is the options for the main http handler
type ServerOpts struct {
	Path             string
	WebListenAddress string
	TLSConfigPath    string
}

// Runs the main web-server
func RunWebServer(opts *ServerOpts, exporter *Exporter, log *logrus.Logger) {
	mux := http.DefaultServeMux

	mux.Handle(opts.Path, exporter.Handler())

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`<html>
            <head><title>Mobserver</title></head>
            <body>
            <h1>Mobserver</h1>
            <p><a href='/metrics'>Metrics</a></p>
            </body>
            </html>`))
		if err != nil {
			log.Errorf("error writing response: %v", err)
		}
	})

	server := &http.Server{
		ReadHeaderTimeout: 2 * time.Second,
		Handler:           mux,
	}
	flags := &web.FlagConfig{
		WebListenAddresses: &[]string{opts.WebListenAddress},
		WebConfigFile:      &opts.TLSConfigPath,
	}
	if err := web.ListenAndServe(server, flags, promlog.New(&promlog.Config{})); err != nil {
		log.Errorf("error starting server: %v", err)
		os.Exit(1)
	}
}
