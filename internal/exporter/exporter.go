package exporter

import (
	"context"
	"fmt"
	"github.com/furiosa-ai/furiosa-metrics-exporter/internal/collector"
	"net/http"
	"time"

	"github.com/furiosa-ai/furiosa-metrics-exporter/internal/config"
	"github.com/furiosa-ai/furiosa-metrics-exporter/internal/pipeline"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

type Exporter struct {
	logger          zerolog.Logger
	collectInterval int
	server          *http.Server
	errChan         chan error
	kubeResSyncChan chan<- struct{}
	pipeline        *pipeline.Pipeline
	deviceRegistry  *collector.DeviceRegistry
	metricFactory   collector.MetricFactory
}

func NewGenericExporter(ctx context.Context, logger zerolog.Logger, cfg *config.Config, deviceRegistry *collector.DeviceRegistry, sampler collector.CoreUtilizationSampler, metricFactory collector.MetricFactory, errChan chan error) (*Exporter, error) {
	kubeResMapper, kubeResSyncChan, err := collector.NewKubeResourcesMapper(ctx, cfg.KubeResourcesLabel)
	if err != nil {
		return nil, err
	}

	newDefaultPipeline := pipeline.NewRegisteredPipeline(deviceRegistry, sampler, metricFactory, kubeResMapper)

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
		errChan:         errChan,
		kubeResSyncChan: kubeResSyncChan,
		pipeline:        newDefaultPipeline,
		deviceRegistry:  deviceRegistry,
		metricFactory:   metricFactory,
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
				// re-enumerate devices, keeping the known set on failure
				added, removed, err := e.deviceRegistry.Sync()
				if err != nil {
					e.logger.Warn().Err(err).Msg("couldn't re-enumerate devices, keeping the previous device set")
				}

				if len(added) > 0 || len(removed) > 0 {
					e.logger.Info().Strs("added", added).Strs("removed", removed).Msg("device set changed")
				}

				// read the driver version once per tick; skipped with no device since there is nothing to read
				if len(e.deviceRegistry.Devices()) > 0 {
					changed, err := e.metricFactory.SyncDriverVersion()
					if err != nil {
						e.logger.Warn().Err(err).Msg("couldn't read driver version, keeping the previous value")
					}

					if changed {
						e.logger.Info().Str("driverVersion", e.metricFactory.DriverVersion()).Msg("driver version changed")
					}
				}

				// trigger kubelet pod resources api
				e.kubeResSyncChan <- struct{}{}

				// collect metrics
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
