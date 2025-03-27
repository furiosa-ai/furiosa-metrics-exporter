package collector

import (
	"errors"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type temperatureCollector struct {
	devices  []smi.Device
	gaugeVec *prometheus.GaugeVec
	nodeName string
}

const (
	ambient = "ambient"
	peak    = "peak"
)

var _ Collector = (*temperatureCollector)(nil)

func NewTemperatureCollector(devices []smi.Device, nodeName string) Collector {
	return &temperatureCollector{
		devices:  devices,
		nodeName: nodeName,
	}
}

func (t *temperatureCollector) Register() {
	t.gaugeVec = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "furiosa_npu_hw_temperature",
		Help: "The current temperature of NPU device",
	},
		[]string{
			arch,
			core,
			device,
			label,
			uuid,
			kubernetesNode,
			kubernetesNamespace,
			kubernetesPod,
			kubernetesContainer,
		})
}

func (t *temperatureCollector) Collect() error {
	metricContainer := make(MetricContainer, 0, len(t.devices))

	errs := make([]error, 0)
	for _, d := range t.devices {
		info, err := getDeviceInfo(d)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		deviceTemperature, err := d.DeviceTemperature()
		if err != nil {
			errs = append(errs, err)
			continue
		}

		metric := newMetric()
		metric[arch] = info.arch
		metric[core] = info.coreLabel
		metric[device] = info.device
		metric[uuid] = info.uuid
		metric[ambient] = deviceTemperature.Ambient()
		metric[peak] = deviceTemperature.SocPeak()

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

func (t *temperatureCollector) postProcess(metrics MetricContainer) error {
	transformed := TransformDeviceMetrics(metrics, false)
	t.gaugeVec.Reset()

	for _, metric := range transformed {
		if value, ok := metric[ambient]; ok {
			t.gaugeVec.With(prometheus.Labels{
				arch:                metric[arch].(string),
				core:                metric[core].(string),
				device:              metric[device].(string),
				label:               ambient,
				uuid:                metric[uuid].(string),
				kubernetesNode:      t.nodeName,
				kubernetesNamespace: metric[kubernetesNamespace].(string),
				kubernetesPod:       metric[kubernetesPod].(string),
				kubernetesContainer: metric[kubernetesContainer].(string),
			}).Set(value.(float64))
		}

		if value, ok := metric[peak]; ok {
			t.gaugeVec.With(prometheus.Labels{
				arch:                metric[arch].(string),
				core:                metric[core].(string),
				device:              metric[device].(string),
				label:               peak,
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
