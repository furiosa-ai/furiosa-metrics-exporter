package cmd

import (
	"context"
	"fmt"
	"github.com/furiosa-ai/furiosa-metrics-exporter/internal/config"
	"github.com/furiosa-ai/furiosa-metrics-exporter/internal/exporter"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/smi"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "furiosa-metrics-exporter",
		Short: "Furiosa Metric Exporter",
		RunE: func(cmd *cobra.Command, args []string) error {
			return start()
		},
	}

	return cmd
}

func start() error {
	ctx, cancelFunc := context.WithCancel(context.Background())
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

	//get config
	config := config.NewDefaultConfig()

	devices, err := smi.ListDevices()
	if err != nil {
		return err
	}

	// Create Exporter
	errChan := make(chan error, 1)
	exporter, err := exporter.NewGenericExporter(config, devices, errChan)
	if err != nil {
		logger.Err(err).Msg("couldn't create exporter")
		return err
	}

	wrappedCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start Exporter
	exporter.Start(wrappedCtx)
	logger.Info().Msg("start event loop")

Loop:
	for {
		select {
		case sig := <-sigChan:
			logger.Err(err).Msg(fmt.Sprintf("signal %d recevied.", sig))
			break Loop
		case errReceived := <-errChan:
			logger.Err(err).Msg(fmt.Sprintf("error %v recevied.", errReceived))
			break Loop
		}
	}

	logger.Info().Msg("stopping metric server")
	err = exporter.Stop()
	if err != nil {
		return err
	}

	return nil
}
