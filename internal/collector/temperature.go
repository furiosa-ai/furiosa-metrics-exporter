package collector

import (
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/smi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type temperature struct {
	devices  []smi.Device
	gaugeVec *prometheus.GaugeVec
}

var _ Collector = (*temperature)(nil)

func NewTemperatureCollector(devices []smi.Device) Collector {
	return &temperature{
		devices: devices,
	}
}

func (t *temperature) Register() {
	t.gaugeVec = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "furiosa_npu_hw_temperature",
		Help: "The current temperature of NPU components",
	},
		[]string{
			"device",
			"label",
		})
}

func (t *temperature) Collect() error {
	// collect is running in a goroutine
	return nil
}

func (t *temperature) Destroy() {
	return
}
