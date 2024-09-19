package collector

import (
	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	utilization = "utilization"
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
	}, []string{arch, core, device, kubernetesNodeName, uuid})
}

func (t *coreUtilizationCollector) Collect() error {
	metricContainer := make(MetricContainer, 0, len(t.devices))
	for _, d := range t.devices {
		metric := make(Metric)

		info, err := getDeviceInfo(d)
		if err != nil {
			return err
		}

		metric[arch] = info.arch
		metric[core] = info.core
		metric[device] = info.device
		metric[kubernetesNodeName] = info.node
		metric[uuid] = info.uuid

		deviceUtilization, err := d.DeviceUtilization()
		if err != nil {
			return err
		}

		metric[utilization] = deviceUtilization

		metricContainer = append(metricContainer, metric)
	}

	return t.postProcess(metricContainer)
}

func (t *coreUtilizationCollector) postProcess(metrics MetricContainer) error {
	for _, metric := range metrics {
		if value, ok := metric[utilization]; ok {
			t.gaugeVec.With(prometheus.Labels{
				arch:               metric[arch].(string),
				core:               metric[core].(string),
				device:             metric[device].(string),
				kubernetesNodeName: metric[kubernetesNodeName].(string),
				label:              utilization,
				uuid:               metric[uuid].(string),
			}).Set(value.(float64))
		}
	}

	return nil
}
