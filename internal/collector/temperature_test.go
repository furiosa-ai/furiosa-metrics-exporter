package collector

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestTempCollector_PostProcessing(t *testing.T) {
	c := &temperatureCollector{}
	c.Register()

	tc := MetricContainer{}
	metric := newMetric()
	metric[arch] = "rngd"
	metric[core] = "0-7"
	metric[device] = "npu0"
	metric[uuid] = uuid
	metric[ambient] = float64(35)
	metric[peak] = float64(39)
	tc = append(tc, metric)
	err := c.postProcess(tc)
	assert.NoError(t, err)

	expected := `
# HELP furiosa_npu_hw_temperature The current temperature of NPU device
# TYPE furiosa_npu_hw_temperature gauge
furiosa_npu_hw_temperature{arch="rngd",core="0-7",device="npu0",driver_version="",firmware_version="",kubernetes_container_name="",kubernetes_namespace_name="",kubernetes_node_name="",kubernetes_pod_name="",label="peak",pci_bus_id="",pert_version="",uuid="uuid"} 39
furiosa_npu_hw_temperature{arch="rngd",core="0-7",device="npu0",driver_version="",firmware_version="",kubernetes_container_name="",kubernetes_namespace_name="",kubernetes_node_name="",kubernetes_pod_name="",label="ambient",pci_bus_id="",pert_version="",uuid="uuid"} 35
`
	err = testutil.GatherAndCompare(prometheus.DefaultGatherer, strings.NewReader(expected), "furiosa_npu_hw_temperature")
	assert.NoError(t, err)
}

func TestTempCollector_Collect(t *testing.T) {
	//TODO: add testcases with device mock
}
