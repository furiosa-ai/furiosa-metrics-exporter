package e2e_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/furiosa-ai/furiosa-metrics-exporter/e2e"
	. "github.com/onsi/ginkgo/v2"
)

const (
	defaultNamespace     = "kube-system"
	defaultHelmChartName = "furiosa-metrics-exporter"

	chartPath = "../deployments/helm"
)

var (
	requiredMetricNames = []string{
		"furiosa_npu_alive",
		"furiosa_npu_error",
		"furiosa_npu_hw_power",
		"furiosa_npu_hw_temperature",
	}

	requiredMetricLabelAttributes = map[string][]string{
		"furiosa_npu_alive": {
			"NO_LABEL_ATTRIBUTE",
		},
		"furiosa_npu_error": {
			"axi_discard_error",
			"axi_doorbell_done",
			"axi_fetch_error",
			"axi_post_error",
			"device_error",
			"pcie_discard_error",
			"pcie_doorbell_done",
			"pcie_fetch_error",
			"pcie_post_error",
		},
		"furiosa_npu_hw_power": {
			"rms",
		},
		"furiosa_npu_hw_temperature": {
			"ambient",
			"peak",
		},
	}
)

var frk e2e.Framework

var valuesYaml = func() string {
	filePath := chartPath + "/values.yaml"
	file, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	return string(file)
}()

func TestFuriosaMetricsExporterE2E(t *testing.T) {
	e2e.GenericRunTestSuiteFunc(t, "'furiosa-metrics-exporter' e2e test")
}

var _ = BeforeSuite(func() {
	fmt.Println("TESTESTS")
	frk = e2e.GenericBeforeSuiteFunc()
	fmt.Println("ASDFASDF")
})

var _ = Describe("'furiosa-metrics-exporter' e2e test", func() {
	Context("Check required metrics exist", Ordered, func() {
		//BeforeAll(e2e.DeployTest(frk, defaultHelmChartName, chartPath, valuesYaml))

		//It("TEST", e2e.DeployTest(frk, defaultHelmChartName, chartPath, valuesYaml))
		//
		fmt.Printf("TEST: %v\n", frk)

		It("Check `Service` is created", func() {})

		It("Wait until `DaemonSet` is ready", func() {})

		It("Get metrics from each pods", func() {})

		It("Validate collected furiosa metrics", func() {})
	})
})
