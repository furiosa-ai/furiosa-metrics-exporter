package collector

import (
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
		})
}

func (t *livenessCollector) Collect() error {
	var metricContainer MetricContainer

	for _, d := range t.devices {
		metric := Metric{}

		info, err := getDeviceInfo(d)
		if err != nil {
			return err
		}

		metric[arch] = info.arch
		metric[core] = info.coreLabel
		metric[device] = info.device
		metric[uuid] = info.uuid
		metric[kubernetesNodeName] = t.nodeName

		value, err := d.Liveness()
		if err != nil {
			return err
		}

		metric[liveness] = value
		metricContainer = append(metricContainer, metric)
	}

	return t.postProcess(metricContainer)
}

func (t *livenessCollector) postProcess(metrics MetricContainer) error {
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
			}).Set(alive)
		}
	}

	return nil
}
