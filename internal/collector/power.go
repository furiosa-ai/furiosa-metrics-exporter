package collector

import (
	"errors"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	rms = "rms"
)

type powerCollector struct {
	devices  []smi.Device
	gaugeVec *prometheus.GaugeVec
	nodeName string
}

var _ Collector = (*powerCollector)(nil)

func NewPowerCollector(devices []smi.Device, nodeName string) Collector {
	return &powerCollector{
		devices:  devices,
		nodeName: nodeName,
	}
}

func (t *powerCollector) Register() {
	t.gaugeVec = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "furiosa_npu_hw_power",
		Help: "The current power of NPU device",
	},
		[]string{
			arch,
			device,
			label,
			core,
			uuid,
			kubernetesNode,
			kubernetesNamespace,
			kubernetesPod,
			kubernetesContainer,
		})
}

func (t *powerCollector) Collect() error {
	metricContainer := make(MetricContainer, 0, len(t.devices))

	errs := make([]error, 0)
	for _, d := range t.devices {
		info, err := getDeviceInfo(d)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		power, err := d.PowerConsumption()
		if err != nil {
			errs = append(errs, err)
			continue
		}

		metric := newMetric()
		metric[arch] = info.arch
		metric[core] = info.coreLabel
		metric[device] = info.device
		metric[uuid] = info.uuid
		metric[rms] = power

		metricContainer = append(metricContainer, metric)
	}

	if err := t.postProcess(metricContainer); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (t *powerCollector) postProcess(metrics MetricContainer) error {
	transformed := TransformDeviceMetrics(metrics, false)
	t.gaugeVec.Reset()

	for _, metric := range transformed {
		if value, ok := metric["rms"]; ok {
			t.gaugeVec.With(prometheus.Labels{
				arch:                metric[arch].(string),
				core:                metric[core].(string),
				device:              metric[device].(string),
				label:               rms,
				uuid:                metric[uuid].(string),
				kubernetesNode:      t.nodeName,
				kubernetesNamespace: metric[kubernetesNamespace].(string),
				kubernetesPod:       metric[kubernetesPod].(string),
				kubernetesContainer: metric[kubernetesContainer].(string),
			}).Set(value.(float64))
		}
	}

	return nil
}
