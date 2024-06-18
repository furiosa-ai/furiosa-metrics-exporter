package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"strings"
	"testing"
)

func TestRegister(t *testing.T) {
	c := &temperature{}
	c.Register()
	c.gaugeVec.WithLabelValues("npu0", "Peak").Set(39000)
	c.gaugeVec.WithLabelValues("npu0", "AMBIENT").Set(35000)

	expected := `
# HELP furiosa_npu_hw_temperature The current temperature of NPU components
# TYPE furiosa_npu_hw_temperature gauge
furiosa_npu_hw_temperature{device="npu0", label="Peak"} 39000.0
furiosa_npu_hw_temperature{device="npu0",label="AMBIENT"} 35000.0
`
	err := testutil.GatherAndCompare(prometheus.DefaultGatherer, strings.NewReader(expected), "furiosa_npu_hw_temperature")
	if err != nil {
		t.Errorf("unexpected collecting result:%s\n", err)
	}
}

//add more tests
