package e2e_test

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/e2e"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	testSuitSpecName     = "'furiosa-metrics-exporter E2E' Test"
	defaultHelmChartName = "furiosa-metrics-exporter"

	chartPath = "../deployments/helm"

	e2eTestPort       = 6254
	e2eTestTargetPort = 6254
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

var (
// valuesObject map[string]any
)

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func composeValues() string {
	imageRegistry := getEnv("E2E_TEST_IMAGE_REGISTRY", "registry.corp.furiosa.ai/furiosa")
	imageName := getEnv("E2E_TEST_IMAGE_NAME", "furiosa-feature-discovery")
	imageTag := getEnv("E2E_TEST_IMAGE_TAG", "latest")

	template := fmt.Sprintf(`namespace: kube-system
daemonSet:
  priorityClassName: system-node-critical

  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
  image:
    repository: %s/%s
    tag: %s
    pullPolicy: IfNotPresent
  resources:
    cpu: 100m
    memory: 64Mi

service:
  type: ClusterIP
  port: %d
  targetPort: %d
  enableScrapAnnotations: true
`, imageRegistry, imageName, imageTag, e2eTestPort, e2eTestTargetPort)
	return template
}

var _ = BeforeSuite(func() { e2e.GenericBeforeSuiteFunc() })

var _ = Describe(testSuitSpecName, func() {
	Context("Check required metrics exist", Ordered, func() {
		BeforeAll(e2e.DeployHelmChart(defaultHelmChartName, chartPath, composeValues()))

		It("Check `Service` is created", checkK8sServiceIsCreated())

		It("Wait until `DaemonSet` is ready", waitUntilDaemonSetIsReady())

		It("Get metrics from each pods and validate it", getMetricFromEachPodsAndValidateIt())

		AfterAll(e2e.DeleteHelmChart())
	})
})

func TestFuriosaMetricsExporterE2E(t *testing.T) {
	e2e.GenericRunTestSuiteFunc(t, testSuitSpecName)
}

func checkK8sServiceIsCreated() func() {
	clientSet := e2e.BackgroundContext().ClientSet

	return func() {
		Eventually(func() int {
			svc, err := clientSet.CoreV1().Services(e2e.BackgroundContext().Namespace).Get(context.TODO(), defaultHelmChartName, metav1.GetOptions{})
			Expect(err).To(BeNil())

			actualPort := int(svc.Spec.Ports[0].Port)
			return actualPort
		}).WithPolling(time.Second * 1).WithTimeout(time.Second * 10).Should(Equal(e2eTestPort))
	}
}

func waitUntilDaemonSetIsReady() func() {
	clientSet := e2e.BackgroundContext().ClientSet

	return func() {
		Eventually(func() bool {
			ds, err := clientSet.AppsV1().DaemonSets(e2e.BackgroundContext().Namespace).Get(context.TODO(), defaultHelmChartName, metav1.GetOptions{})
			if err != nil {
				return false
			}

			currentNumberScheduled := ds.Status.CurrentNumberScheduled
			desiredNumberScheduled := ds.Status.DesiredNumberScheduled
			numberReady := ds.Status.NumberReady

			return currentNumberScheduled == desiredNumberScheduled && desiredNumberScheduled == numberReady
		}).WithPolling(time.Second * 1).WithTimeout(time.Second * 30).Should(BeTrue())
	}
}

func getMetricFromEachPodsAndValidateIt() func() {
	clientSet := e2e.BackgroundContext().ClientSet

	return func() {
		helmReleaseName := e2e.BackgroundContext().HelmChart.ReleaseName
		labelSelector := fmt.Sprintf("app.kubernetes.io/name=%s,app.kubernetes.io/instance=%s", defaultHelmChartName, helmReleaseName)
		listOptions := metav1.ListOptions{LabelSelector: labelSelector}

		podList, err := clientSet.CoreV1().Pods(e2e.BackgroundContext().Namespace).List(context.TODO(), listOptions)
		Expect(err).To(BeNil())

		nodeNameToLineByLineMetricsMap := make(map[string][]string)

		for _, pod := range podList.Items {
			Eventually(func() bool {
				res := clientSet.CoreV1().RESTClient().Get().
					Namespace(pod.Namespace).
					Resource("pods").
					Name(fmt.Sprintf("%s:%d", pod.Name, e2eTestTargetPort)).
					SubResource("proxy").
					Suffix("metrics").
					Do(context.TODO())

				resBody, err := res.Raw()
				if err != nil {
					return false
				}

				requiredMetricNameCheckMap := make(map[string]struct{})

				// furiosa_npu_alive, furiosa_npu_error,furiosa_npu_hw_power,furiosa_npu_hw_temperature
				lineByLineMetrics := strings.Split(string(resBody), "\n")
				requiredMetrics := make([]string, 0, len(lineByLineMetrics))
				for _, lineByLineMetric := range lineByLineMetrics {
					idx := strings.Index(lineByLineMetric, "{")
					if idx == -1 {
						continue
					}

					metricName := lineByLineMetric[:idx]

					if slices.Contains(requiredMetricNames, metricName) {
						requiredMetricNameCheckMap[metricName] = struct{}{}
						requiredMetrics = append(requiredMetrics, lineByLineMetric)
					}
				}

				if len(requiredMetricNames) == len(requiredMetricNameCheckMap) {
					nodeNameToLineByLineMetricsMap[pod.Spec.NodeName] = requiredMetrics

					return true
				}

				return false
			}).WithPolling(time.Second * 15).WithTimeout(time.Second * 300).Should(BeTrue())
		}

		for nodeName, lineByLineMetrics := range nodeNameToLineByLineMetricsMap {
			checklistByDevice := make(map[string]map[string]map[string]struct{})
			for _, lineByLineMetric := range lineByLineMetrics {
				metricName, labels, err := parseMetricLineDataString(lineByLineMetric)
				Expect(err).To(BeNil())
				Expect(labels["kubernetes_node_name"]).To(Equal(nodeName))

				device := labels["device"]
				if _, ok := checklistByDevice[device]; !ok {
					checklistByDevice[device] = make(map[string]map[string]struct{})
					checklistByDevice[device]["furiosa_npu_alive"] = make(map[string]struct{})
					checklistByDevice[device]["furiosa_npu_error"] = make(map[string]struct{})
					checklistByDevice[device]["furiosa_npu_hw_temperature"] = make(map[string]struct{})
					checklistByDevice[device]["furiosa_npu_hw_power"] = make(map[string]struct{})
				}

				switch metricName {
				case "furiosa_npu_alive":
					checklistByDevice[device][metricName]["NO_LABEL_ATTRIBUTE"] = struct{}{}

				case "furiosa_npu_error":
					attribute := labels["label"]
					checklistByDevice[device][metricName][attribute] = struct{}{}

				case "furiosa_npu_hw_temperature":
					attribute := labels["label"]
					checklistByDevice[device][metricName][attribute] = struct{}{}

				case "furiosa_npu_hw_power":
					attribute := labels["label"]
					checklistByDevice[device][metricName][attribute] = struct{}{}
				}
			}

			for _, metricsWithLabelAttributesMap := range checklistByDevice {
				for metricName, labelAttributesMap := range metricsWithLabelAttributesMap {
					labelAttributes := make([]string, 0, len(labelAttributesMap))
					for key := range labelAttributesMap {
						labelAttributes = append(labelAttributes, key)
					}

					Expect(labelAttributes).Should(ContainElements(requiredMetricLabelAttributes[metricName]))
				}
			}
		}
	}
}

func parseMetricLineDataString(metric string) (string, map[string]string, error) {
	braceHead, braceTail := strings.Index(metric, "{"), strings.LastIndex(metric, "}")
	if braceHead == -1 || braceTail == -1 {
		return "", nil, fmt.Errorf("unable to parse metric labels because '{' or '}' is missing")
	}

	metricName := metric[:braceHead]

	labelsSection := metric[braceHead+1 : braceTail]
	keyValuePairLabels := strings.Split(labelsSection, ",")

	labels := make(map[string]string)
	for _, keyValuePair := range keyValuePairLabels {
		keyValuePairSlice := strings.Split(keyValuePair, "=")
		key, value := keyValuePairSlice[0], keyValuePairSlice[1][1:len(keyValuePairSlice[1])-1]

		labels[key] = value
	}

	return metricName, labels, nil
}
