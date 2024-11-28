package collector

import (
	"errors"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
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

var errorLabels = []string{
	axiPostError,
	axiFetchError,
	axiDiscardError,
	axiDoorbellDone,
	pciePostError,
	pcieFetchError,
	pcieDiscardError,
	pcieDoorbellDone,
	deviceError,
}

type errorCollector struct {
	devices  []smi.Device
	gaugeVec *prometheus.GaugeVec
	nodeName string
}

var _ Collector = (*errorCollector)(nil)

func NewErrorCollector(devices []smi.Device, nodeName string) Collector {
	return &errorCollector{
		devices:  devices,
		nodeName: nodeName,
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
	metricContainer := make(MetricContainer, 0, len(t.devices))

	errs := make([]error, 0)
	for _, d := range t.devices {
		metric := Metric{}

		info, err := getDeviceInfo(d)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		metric[arch] = info.arch
		metric[device] = info.device
		metric[uuid] = info.uuid
		metric[core] = info.coreLabel
		metric[kubernetesNodeName] = t.nodeName

		errorInfo, err := d.DeviceErrorInfo()
		if err != nil {
			errs = append(errs, err)
			continue
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

	if err := t.postProcess(metricContainer); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (t *errorCollector) postProcess(metrics MetricContainer) error {
	t.gaugeVec.Reset()

	for _, metric := range metrics {
		for _, errorLabel := range errorLabels {
			if val, ok := metric[errorLabel]; ok {
				t.gaugeVec.With(prometheus.Labels{
					arch:               metric[arch].(string),
					device:             metric[device].(string),
					uuid:               metric[uuid].(string),
					label:              errorLabel,
					core:               metric[core].(string),
					kubernetesNodeName: metric[kubernetesNodeName].(string),
				}).Set(val.(float64))
			}
		}
	}

	return nil
}
