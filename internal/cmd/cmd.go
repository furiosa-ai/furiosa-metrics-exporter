package cmd

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

			if port, err := cmd.Flags().GetInt("port"); err != nil {
				return err
			} else {
				cfg.SetPort(port)
			}

			if interval, err := cmd.Flags().GetInt("interval"); err != nil {
				return err
			} else {
				cfg.SetInterval(interval)
			}

			if nodeName, err := cmd.Flags().GetString("node-name"); err != nil {
				return err
			} else {
				cfg.SetNodeName(nodeName)
			}

			return Run(cmd.Context(), cfg)
		},
	}

	cmd.Flags().Int("port", 0, "Port number used for metrics server")
	cmd.Flags().Int("interval", 0, "Collection interval value in second")
	cmd.Flags().String("node-name", "", "Node name of the current execution environment")

	if err := cmd.MarkFlagRequired("port"); err != nil {
		panic(err)
	}

	if err := cmd.MarkFlagRequired("interval"); err != nil {
		panic(err)
	}

	return cmd
}

func Run(ctx context.Context, cfg *config.Config) error {
	ctx, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc()

	// create core loop logger
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("subject", "core_loop").Logger()
	_ = logger.WithContext(ctx)

	//os signal listener
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

	// Create Exporter
	errChan := make(chan error, 1)
	metricsExporter, err := exporter.NewGenericExporter(cfg, devices, errChan)
	if err != nil {
		logger.Err(err).Msg("couldn't create exporter")
		return err
	}

	// TODO(@hoony9x) 위에 cancelFunc 있는데 왜 또 있죠?
	//	wrappedCtx, cancel := context.WithCancel(ctx)
	//	defer cancel()

	// Start Exporter
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

	logger.Info().Msg("stopping metric server")
	err = metricsExporter.Stop()
	if err != nil {
		return err
	}

	return nil
}
