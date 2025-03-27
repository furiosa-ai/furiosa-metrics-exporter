package exporter

import (
	"context"
	"fmt"
	"github.com/furiosa-ai/furiosa-metrics-exporter/internal/collector"
	"net/http"
	"time"

	"github.com/furiosa-ai/furiosa-metrics-exporter/internal/config"
	"github.com/furiosa-ai/furiosa-metrics-exporter/internal/pipeline"
	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

type Exporter struct {
	logger          zerolog.Logger
	collectInterval int
	server          *http.Server
	errChan         chan error
	pipeline        *pipeline.Pipeline
}

func NewGenericExporter(logger zerolog.Logger, cfg *config.Config, devices []smi.Device, errChan chan error) (*Exporter, error) {
	exporter := Exporter{
		logger:          logger,
		collectInterval: cfg.Interval,
		server: &http.Server{
			Addr: fmt.Sprintf(":%d", cfg.Port),
			Handler: func() http.Handler {
				// build Webserver
				mux := http.NewServeMux()
				mux.Handle("/metrics", promhttp.Handler())

				return mux
			}(),
		},
		errChan:  errChan,
		pipeline: pipeline.NewRegisteredPipeline(devices, cfg.NodeName),
	}

	return &exporter, nil
}

func (e *Exporter) Start(ctx context.Context) {
	//run pipeline
	go func() {
		tick := time.NewTicker(time.Second * time.Duration(e.collectInterval))

		// When panic happens, send error to the `errChan`, and call `ctx.Done()` to exit.
		defer func() {
			if r := recover(); r != nil {
				e.errChan <- fmt.Errorf("recovered from panic: %v", r)
				ctx.Done()
			}
		}()

		for {
			select {
			case <-tick.C:

				collector.SyncPodInfoCache()
				for _, err := range e.pipeline.Collect() {
					e.logger.Err(err).Msg(fmt.Sprintf("error %v received from pipeline collector", err))
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

func (e *Exporter) Stop(ctx context.Context) error {
	//stop web server
	err := e.server.Shutdown(ctx)
	if err != nil {
		return err
	}

	return nil
}
