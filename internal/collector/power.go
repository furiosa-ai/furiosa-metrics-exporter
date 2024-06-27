package collector

import (
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/smi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	rms = "rms"
)

type powerCollector struct {
	devices  []smi.Device
	gaugeVec *prometheus.GaugeVec
}

var _ Collector = (*powerCollector)(nil)

func NewPowerCollector(devices []smi.Device) Collector {
	return &powerCollector{
		devices: devices,
	}
}

func (t *powerCollector) Register() {
	t.gaugeVec = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "furiosa_npu_hw_power",
		Help: "The current power of NPU components",
	},
		[]string{
			arch,
			device,
			label,
			core,
			kubernetesNodeName,
			uuid,
		})
}

func (t *powerCollector) Collect() error {
	var metricContainer MetricContainer

	for _, d := range t.devices {
		metric := Metric{}

		info, err := getDeviceInfo(d)
		if err != nil {
			return err
		}

		metric[arch] = info.arch
		metric[device] = info.device
		metric[uuid] = info.uuid
		metric[core] = info.core
		metric[kubernetesNodeName] = info.node

		power, err := d.PowerConsumption()
		if err != nil {
			return err
		}

		metric[rms] = power
		metricContainer = append(metricContainer, metric)
	}

	return t.postProcess(metricContainer)
}

func (t *powerCollector) postProcess(metrics MetricContainer) error {
	for _, metric := range metrics {
		if value, ok := metric["rms"]; ok {

			t.gaugeVec.With(prometheus.Labels{
				arch:               metric[arch].(string),
				core:               metric[core].(string),
				device:             metric[device].(string),
				kubernetesNodeName: metric[kubernetesNodeName].(string),
				label:              rms,
				uuid:               metric[uuid].(string),
			}).Set(value.(float64))
		}
	}

	return nil
}
