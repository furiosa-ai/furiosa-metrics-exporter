package collector

import (
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/smi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	axiPostError     = "axi_post_error"
	axiFetchError    = "axi_fetch_error"
	axiDiscardError  = "axi_discard_error"
	axiDoorbellDone  = "axi_doorbell_done"
	pciePostError    = "pcie_post_error"
	pcieFetchError   = "pcie_fetch_error"
	pcieDiscardError = "pcie_discard_error"
	pcieDoorbellDone = "pcie_doorbell_done"
	deviceError      = "device_error"
)

type errorCollector struct {
	devices  []smi.Device
	gaugeVec *prometheus.GaugeVec
}

var _ Collector = (*errorCollector)(nil)

func NewErrorCollector(devices []smi.Device) Collector {
	return &errorCollector{
		devices: devices,
	}
}

func (t *errorCollector) Register() {
	t.gaugeVec = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "furiosa_npu_error",
		Help: "The current active error counts of NPU device",
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

func (t *errorCollector) Collect() error {
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

		errorInfo, err := d.DeviceErrorInfo()
		if err != nil {
			return err
		}

		metric[axiPostError] = float64(errorInfo.AxiPostErrorCount())
		metric[axiFetchError] = float64(errorInfo.AxiFetchErrorCount())
		metric[axiDiscardError] = float64(errorInfo.AxiDiscardErrorCount())
		metric[axiDoorbellDone] = float64(errorInfo.AxiDoorbellErrorCount())
		metric[pciePostError] = float64(errorInfo.PciePostErrorCount())
		metric[pcieFetchError] = float64(errorInfo.PcieFetchErrorCount())
		metric[pcieDiscardError] = float64(errorInfo.PcieDiscardErrorCount())
		metric[pcieDoorbellDone] = float64(errorInfo.PcieDoorbellErrorCount())
		metric[deviceError] = float64(errorInfo.DeviceErrorCount())
		metricContainer = append(metricContainer, metric)
	}

	return t.postProcess(metricContainer)
}

func (t *errorCollector) postProcess(metrics MetricContainer) error {
	for _, metric := range metrics {
		if val, ok := metric[axiPostError]; ok {
			t.gaugeVec.With(prometheus.Labels{
				arch:               metric[arch].(string),
				device:             metric[device].(string),
				uuid:               metric[uuid].(string),
				label:              axiPostError,
				core:               metric[core].(string),
				kubernetesNodeName: metric[kubernetesNodeName].(string),
			}).Set(val.(float64))
		}

		if val, ok := metric[axiFetchError]; ok {
			t.gaugeVec.With(prometheus.Labels{
				arch:               metric[arch].(string),
				device:             metric[device].(string),
				uuid:               metric[uuid].(string),
				label:              axiFetchError,
				core:               metric[core].(string),
				kubernetesNodeName: metric[kubernetesNodeName].(string),
			}).Set(val.(float64))
		}

		if val, ok := metric[axiDiscardError]; ok {
			t.gaugeVec.With(prometheus.Labels{
				arch:               metric[arch].(string),
				device:             metric[device].(string),
				uuid:               metric[uuid].(string),
				label:              axiDiscardError,
				core:               metric[core].(string),
				kubernetesNodeName: metric[kubernetesNodeName].(string),
			}).Set(val.(float64))
		}

		if val, ok := metric[axiDoorbellDone]; ok {
			t.gaugeVec.With(prometheus.Labels{
				arch:               metric[arch].(string),
				device:             metric[device].(string),
				uuid:               metric[uuid].(string),
				label:              axiDoorbellDone,
				core:               metric[core].(string),
				kubernetesNodeName: metric[kubernetesNodeName].(string),
			}).Set(val.(float64))
		}

		if val, ok := metric[pciePostError]; ok {
			t.gaugeVec.With(prometheus.Labels{
				arch:               metric[arch].(string),
				device:             metric[device].(string),
				uuid:               metric[uuid].(string),
				label:              pciePostError,
				core:               metric[core].(string),
				kubernetesNodeName: metric[kubernetesNodeName].(string),
			}).Set(val.(float64))
		}

		if val, ok := metric[pcieFetchError]; ok {
			t.gaugeVec.With(prometheus.Labels{
				arch:               metric[arch].(string),
				device:             metric[device].(string),
				uuid:               metric[uuid].(string),
				label:              pcieFetchError,
				core:               metric[core].(string),
				kubernetesNodeName: metric[kubernetesNodeName].(string),
			}).Set(val.(float64))
		}

		if val, ok := metric[pcieDiscardError]; ok {
			t.gaugeVec.With(prometheus.Labels{
				arch:               metric[arch].(string),
				device:             metric[device].(string),
				uuid:               metric[uuid].(string),
				label:              pcieDiscardError,
				core:               metric[core].(string),
				kubernetesNodeName: metric[kubernetesNodeName].(string),
			}).Set(val.(float64))
		}

		if val, ok := metric[pcieDoorbellDone]; ok {
			t.gaugeVec.With(prometheus.Labels{
				arch:               metric[arch].(string),
				device:             metric[device].(string),
				uuid:               metric[uuid].(string),
				label:              pcieDoorbellDone,
				core:               metric[core].(string),
				kubernetesNodeName: metric[kubernetesNodeName].(string),
			}).Set(val.(float64))
		}

		if val, ok := metric[deviceError]; ok {
			t.gaugeVec.With(prometheus.Labels{
				arch:               metric[arch].(string),
				device:             metric[device].(string),
				uuid:               metric[uuid].(string),
				label:              deviceError,
				core:               metric[core].(string),
				kubernetesNodeName: metric[kubernetesNodeName].(string),
			}).Set(val.(float64))
		}
	}

	return nil
}
