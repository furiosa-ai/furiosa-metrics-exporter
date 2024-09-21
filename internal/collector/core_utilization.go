package collector

import (
	"fmt"
	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	peUtilization = "peUtilization"
)

type coreUtilizationCollector struct {
	devices  []smi.Device
	gaugeVec *prometheus.GaugeVec
}

var _ Collector = (*coreUtilizationCollector)(nil)

func NewCoreUtilizationCollector(devices []smi.Device) Collector {
	return &coreUtilizationCollector{
		devices: devices,
	}
}

func (t *coreUtilizationCollector) Register() {
	t.gaugeVec = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "furiosa_npu_core_utilization",
		Help: "The current core utilization of NPU device",
	},
		[]string{
			arch,
			device,
			core,
			kubernetesNodeName,
			uuid,
		})
}

func (t *coreUtilizationCollector) Collect() error {
	var metricContainer MetricContainer

	for _, d := range t.devices {
		info, err := getDeviceInfo(d)
		if err != nil {
			return err
		}

		deviceUtilization, err := d.DeviceUtilization()
		if err != nil {
			return err
		}

		utilization := deviceUtilization.PeUtilization()
		for _, pe := range utilization {
			metric := Metric{
				arch:               info.arch,
				core:               fmt.Sprintf("%d", pe.Core()),
				device:             info.device,
				kubernetesNodeName: info.node,
				uuid:               info.uuid,
				peUtilization:      pe.PeUsagePercentage(),
			}
			metricContainer = append(metricContainer, metric)
		}
	}

	return t.postProcess(metricContainer)
}

func (t *coreUtilizationCollector) postProcess(metrics MetricContainer) error {
	for _, metric := range metrics {
		if value, ok := metric[peUtilization]; ok {
			t.gaugeVec.With(prometheus.Labels{
				arch:               metric[arch].(string),
				core:               metric[core].(string),
				device:             metric[device].(string),
				kubernetesNodeName: metric[kubernetesNodeName].(string),
				uuid:               metric[uuid].(string),
			}).Set(value.(float64))
		}
	}

	return nil
}
