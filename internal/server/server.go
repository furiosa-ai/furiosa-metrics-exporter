package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/furiosa-ai/furiosa-metrics-exporter/internal/config"
	"github.com/furiosa-ai/furiosa-metrics-exporter/internal/exporter"
	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "furiosa-metrics-exporter",
		Short: "Furiosa Metric Exporter",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.NewDefaultConfig()
			cfg.SetFromFlags(cmd)

			return Run(cmd.Context(), cfg)
		},
	}

	return cmd
}

func Run(ctx context.Context, cfg *config.Config) error {
	ctx, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc()

	// Create core loop logger
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("subject", "core_loop").Logger()
	_ = logger.WithContext(ctx)

	// OS signal listener
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	defer func() {
		logger.Info().Msg("closing channels")
		signal.Stop(sigChan)
		close(sigChan)
	}()

	devices, err := smi.ListDevices()
	if err != nil {
		return err
	}

	// Create Metrics Exporter
	errChan := make(chan error, 1)
	metricsExporter, err := exporter.NewGenericExporter(cfg, devices, errChan)
	if err != nil {
		logger.Err(err).Msg("couldn't create metrics exporter")
		return err
	}

	// TODO(@hoony9x) 위에 cancelFunc 있는데 왜 또 있죠?
	//	wrappedCtx, cancel := context.WithCancel(ctx)
	//	defer cancel()

	// Start Metrics Exporter
	metricsExporter.Start(ctx)
	logger.Info().Msg("start event loop")

Loop:
	for {
		select {
		case sig := <-sigChan:
			logger.Err(err).Msg(fmt.Sprintf("signal %d received.", sig))
			break Loop

		case errReceived := <-errChan:
			logger.Err(err).Msg(fmt.Sprintf("error %v received.", errReceived))
			break Loop
		}
	}

	logger.Info().Msg("stop metrics server")
	err = metricsExporter.Stop()
	if err != nil {
		return err
	}

	return nil
}
