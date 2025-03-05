package collector

import (
	"errors"

	"github.com/furiosa-ai/furiosa-metrics-exporter/internal/kubernetes"
	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	liveness = "liveness"
)

type livenessCollector struct {
	devices  []smi.Device
	gaugeVec *prometheus.GaugeVec
	nodeName string
}

var _ Collector = (*livenessCollector)(nil)

func NewLivenessCollector(devices []smi.Device, nodeName string) Collector {
	return &livenessCollector{
		devices:  devices,
		nodeName: nodeName,
	}
}

func (t *livenessCollector) Register() {
	t.gaugeVec = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "furiosa_npu_alive",
		Help: "The liveness of NPU device",
	},
		[]string{
			arch,
			core,
			device,
			kubernetesNodeName,
			uuid,
			pod,
		})
}

func (t *livenessCollector) Collect(devicePodMap map[string][]kubernetes.PodInfo) error {
	metricContainer := make(MetricContainer, 0, len(t.devices))

	errs := make([]error, 0)
	for _, d := range t.devices {
		metric := Metric{}

		info, err := getDeviceInfo(d)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		value, err := d.Liveness()
		if err != nil {
			errs = append(errs, err)
			continue
		}

		if podInfos, ok := devicePodMap[info.uuid]; ok {
			if !(len(podInfos) == 1 && len(podInfos[0].AllocatedPE) == 8) { // Partitioned device allocation case. Add original card metric.
				metric[arch] = info.arch
				metric[core] = info.coreLabel
				metric[device] = info.device
				metric[uuid] = info.uuid
				metric[kubernetesNodeName] = t.nodeName
				metric[pod] = ""
				metric[liveness] = value

				metricContainer = append(metricContainer, metric)
			}

			for _, podInfo := range podInfos {
				metric := Metric{}
				metric[arch] = info.arch
				metric[core] = podInfo.CoreLabel
				metric[device] = info.device
				metric[uuid] = info.uuid
				metric[kubernetesNodeName] = t.nodeName
				metric[pod] = podInfo.Name
				metric[liveness] = value

				metricContainer = append(metricContainer, metric)
			}
		} else { // Non-allocated device case
			metric[arch] = info.arch
			metric[core] = info.coreLabel
			metric[device] = info.device
			metric[uuid] = info.uuid
			metric[kubernetesNodeName] = t.nodeName
			metric[pod] = ""
			metric[liveness] = value

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

func (t *livenessCollector) postProcess(metrics MetricContainer) error {
	t.gaugeVec.Reset()

	for _, metric := range metrics {
		if value, ok := metric[liveness]; ok {
			var alive float64
			if value.(bool) {
				alive = 1
			} else {
				alive = 0
			}

			t.gaugeVec.With(prometheus.Labels{
				arch:               metric[arch].(string),
				core:               metric[core].(string),
				device:             metric[device].(string),
				uuid:               metric[uuid].(string),
				kubernetesNodeName: metric[kubernetesNodeName].(string),
				pod:                metric[pod].(string),
			}).Set(alive)

		}
	}

	return nil
}
