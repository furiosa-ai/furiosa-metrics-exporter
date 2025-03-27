package collector

import (
	"errors"
	"strconv"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	peFrequency = "coreFrequency"
)

type coreFrequencyCollector struct {
	devices  []smi.Device
	gaugeVec *prometheus.GaugeVec
	nodeName string
}

var _ Collector = (*coreFrequencyCollector)(nil)

func NewCoreFrequencyCollector(devices []smi.Device, nodeName string) Collector {
	return &coreFrequencyCollector{
		devices:  devices,
		nodeName: nodeName,
	}
}

func (t *coreFrequencyCollector) Register() {
	t.gaugeVec = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "furiosa_npu_core_frequency",
		Help: "The current core frequency of NPU device (MHz)",
	},
		[]string{
			arch,
			device,
			core,
			uuid,
			kubernetesNode,
			kubernetesNamespace,
			kubernetesPod,
			kubernetesContainer,
		})
}

func (t *coreFrequencyCollector) Collect() error {
	metricContainer := make(MetricContainer, 0, len(t.devices))

	errs := make([]error, 0)
	for _, d := range t.devices {
		info, err := getDeviceInfo(d)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		coreFrequency, err := d.CoreFrequency()
		if err != nil {
			errs = append(errs, err)
			continue
		}

		frequency := coreFrequency.PeFrequency()
		for _, pe := range frequency {
			metric := newMetric()
			metric[arch] = info.arch
			metric[core] = strconv.Itoa(int(pe.Core()))
			metric[device] = info.device
			metric[uuid] = info.uuid
			metric[peFrequency] = pe.Frequency()

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

func (t *coreFrequencyCollector) postProcess(metrics MetricContainer) error {
	transformed := TransformDeviceMetrics(metrics, true)
	t.gaugeVec.Reset()

	for _, metric := range transformed {
		if value, ok := metric[peFrequency]; ok {

			t.gaugeVec.With(prometheus.Labels{
				arch:                metric[arch].(string),
				core:                metric[core].(string),
				device:              metric[device].(string),
				uuid:                metric[uuid].(string),
				kubernetesNode:      t.nodeName,
				kubernetesNamespace: metric[kubernetesNamespace].(string),
				kubernetesPod:       metric[kubernetesPod].(string),
				kubernetesContainer: metric[kubernetesContainer].(string),
			}).Set(float64(value.(uint32)))
		}
	}

	return nil
}
