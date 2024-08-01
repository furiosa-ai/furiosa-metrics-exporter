package exporter

import (
	"context"
	"github.com/furiosa-ai/furiosa-metrics-exporter/internal/config"
	"github.com/furiosa-ai/furiosa-metrics-exporter/internal/pipeline"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/smi"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"time"
)

type Exporter struct {
	conf     *config.Config
	server   *http.Server
	errChan  chan error
	pipeline *pipeline.Pipeline
}

func NewGenericExporter(conf *config.Config, devices []smi.Device, errChan chan error) (*Exporter, error) {
	exporter := Exporter{
		conf:     conf,
		errChan:  errChan,
		pipeline: pipeline.NewRegisteredPipeline(devices),
	}

	// build Webserver
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	exporter.server = &http.Server{
		Addr:    ":6254",
		Handler: mux,
	}

	return &exporter, nil
}

func (e *Exporter) Start(ctx context.Context) {
	//run pipeline
	go func() {
		tick := time.NewTicker(time.Second * time.Duration(e.conf.Interval))

		for {
			select {
			case <-tick.C:
				err := e.pipeline.Collect()
				if err != nil {
					e.errChan <- err
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	//start web server
	go func() {
		err := e.server.ListenAndServe()
		if err != nil {
			e.errChan <- err
			return
		}
	}()
}

func (e *Exporter) Stop() error {
	//stop web server
	err := e.server.Shutdown(context.TODO())
	if err != nil {
		return err
	}
	return nil
}
