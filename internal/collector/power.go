package collector

import (
	"errors"

	"github.com/furiosa-ai/furiosa-metrics-exporter/internal/kubernetes"
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
			kubernetesNodeName,
			uuid,
			pod,
		})
}

func (t *powerCollector) Collect(devicePodMap map[string][]kubernetes.PodInfo) error {
	metricContainer := make(MetricContainer, 0, len(t.devices))

	errs := make([]error, 0)
	for _, d := range t.devices {
		metric := Metric{}

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

		if PodInfos, ok := devicePodMap[info.uuid]; ok {
			if !(len(PodInfos) == 1 && len(PodInfos[0].AllocatedPE) == 8) { // Partitioned device allocation case. Add original card metric.
				metric[arch] = info.arch
				metric[device] = info.device
				metric[uuid] = info.uuid
				metric[core] = info.coreLabel
				metric[kubernetesNodeName] = t.nodeName
				metric[pod] = ""
				metric[rms] = power

				metricContainer = append(metricContainer, metric)
			}

			for _, podInfo := range PodInfos {
				metric := Metric{}
				metric[arch] = info.arch
				metric[core] = podInfo.CoreLabel
				metric[device] = info.device
				metric[uuid] = info.uuid
				metric[kubernetesNodeName] = t.nodeName
				metric[pod] = podInfo.Name
				metric[rms] = power

				metricContainer = append(metricContainer, metric)
			}
		} else { // Non-allocated device case
			metric[arch] = info.arch
			metric[device] = info.device
			metric[uuid] = info.uuid
			metric[core] = info.coreLabel
			metric[kubernetesNodeName] = t.nodeName
			metric[pod] = ""
			metric[rms] = power

			metricContainer = append(metricContainer, metric)
		}
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
	t.gaugeVec.Reset()

	for _, metric := range metrics {
		if value, ok := metric["rms"]; ok {
			t.gaugeVec.With(prometheus.Labels{
				arch:               metric[arch].(string),
				core:               metric[core].(string),
				device:             metric[device].(string),
				kubernetesNodeName: metric[kubernetesNodeName].(string),
				label:              rms,
				uuid:               metric[uuid].(string),
				pod:                metric[pod].(string),
			}).Set(value.(float64))
		}
	}

	return nil
}
