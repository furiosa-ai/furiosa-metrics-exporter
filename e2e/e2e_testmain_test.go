package e2e_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bradfitz/iter"
	"github.com/google/uuid"
	helmclient "github.com/mittwald/go-helm-client"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

const (
	defaultKubeConfigPath = ".kube/config"
	defaultNamespace      = "kube-system"
	defaultHelmChartName  = "furiosa-metrics-exporter"

	k8sClientCtxKey                     = "K8S_CLIENT"
	helmClientCtxKey                    = "HELM_CLIENT"
	helmReleaseNameCtxKey               = "HELM_RELEASE_NAME"
	nodeNameToLineByLineMetricMapCtxKey = "NODE_NAME_TO_LINE_BY_LINE_MAP"
	portNumberCtxKey                    = "PORT_NUMBER"
)

var (
	testenv env.Environment = nil

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

func TestMain(m *testing.M) {
	kubeConfigPath := os.Getenv("KUBECONFIG")
	if kubeConfigPath == "" {
		fmt.Println("KUBECONFIG environment variable not set. Use default path '~/.kube/config'")

		homePath, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}

		kubeConfigPath = homePath + "/" + defaultKubeConfigPath
	}

	var helmReleaseName string
	{
		UUID, _ := uuid.NewUUID()
		helmReleaseName = fmt.Sprintf("e2e-test-%s", UUID.String()[0:8])
	}

	var absolutePath string
	{
		relativePath := "../deployments/helm"
		absPath, err := filepath.Abs(relativePath)
		if err != nil {
			panic(err)
		}

		absolutePath = absPath
	}

	var valuesYaml string
	{
		filePath := absolutePath + "/values.yaml"
		file, err := os.ReadFile(filePath)
		if err != nil {
			panic(err)
		}

		valuesYaml = string(file)
	}

	testenv = env.New()

	testenv.Setup(
		// Setup K8s Client
		func(ctx context.Context, config *envconf.Config) (context.Context, error) {
			config.WithKubeconfigFile(kubeConfigPath)

			return ctx, nil
		},
		// Setup K8s ClientSet and Helm Client
		func(ctx context.Context, config *envconf.Config) (context.Context, error) {
			restConfig, err := clientcmd.BuildConfigFromFlags("", config.KubeconfigFile())
			if err != nil {
				return ctx, err
			}

			k8sClient, err := kubernetes.NewForConfig(restConfig)
			if err != nil {
				return ctx, err
			}

			helmClientOptions := &helmclient.RestConfClientOptions{
				Options: &helmclient.Options{
					Namespace: defaultNamespace,
				},
				RestConfig: restConfig,
			}
			helmClient, err := helmclient.NewClientFromRestConf(helmClientOptions)
			if err != nil {
				return ctx, err
			}

			ctx = context.WithValue(ctx, k8sClientCtxKey, k8sClient)
			ctx = context.WithValue(ctx, helmClientCtxKey, helmClient)

			return ctx, nil
		},
		// Install Helm Client
		func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
			helmClient, _ := ctx.Value(helmClientCtxKey).(helmclient.Client)
			helmChartSpec := &helmclient.ChartSpec{
				ReleaseName:     helmReleaseName,
				ChartName:       absolutePath,
				Namespace:       defaultNamespace,
				CreateNamespace: false,
				Wait:            true,
				Timeout:         1 * time.Minute,
				CleanupOnFail:   false,
				ValuesYaml:      valuesYaml,
			}

			if _, err := helmClient.InstallChart(ctx, helmChartSpec, nil); err != nil {
				return ctx, err
			}

			ctx = context.WithValue(ctx, helmReleaseNameCtxKey, helmReleaseName)

			return ctx, nil
		},
	)

	testenv.Finish(
		// Uninstall Helm Client
		func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
			helmClient, _ := ctx.Value(helmClientCtxKey).(helmclient.Client)
			if err := helmClient.UninstallReleaseByName(helmReleaseName); err != nil {
				return ctx, err
			}

			return ctx, nil
		},
	)

	exitCode := testenv.Run(m)
	os.Exit(exitCode)
}

func TestFuriosaMetricsExporter(t *testing.T) {
	f := features.New("Check required metrics exist").
		Assess("Check `Service` is created", func(ctx context.Context, subT *testing.T, config *envconf.Config) context.Context {
			service := corev1.Service{}
			if err := config.Client().Resources().Get(ctx, defaultHelmChartName, defaultNamespace, &service); err != nil {
				subT.Fatal(err)
			}

			portNumber := service.Spec.Ports[0].Port

			ctx = context.WithValue(ctx, portNumberCtxKey, int(portNumber))
			return ctx
		}).
		Assess("Wait until `DaemonSet` is ready", func(ctx context.Context, subT *testing.T, config *envconf.Config) context.Context {
			daemonset := appsv1.DaemonSet{}
			if err := config.Client().Resources().Get(ctx, defaultHelmChartName, defaultNamespace, &daemonset); err != nil {
				subT.Fatal(err)
			}

			waitCondition := conditions.New(config.Client().Resources()).DaemonSetReady(&daemonset)
			waitTimeout := wait.WithTimeout(time.Minute * 1)
			waitInterval := wait.WithInterval(time.Second * 10)

			if err := wait.For(waitCondition, waitTimeout, waitInterval); err != nil {
				subT.Fatal(err)
			}

			return ctx
		}).
		Assess("Get metrics from each pods", func(ctx context.Context, subT *testing.T, config *envconf.Config) context.Context {
			helmReleaseName, _ := ctx.Value(helmReleaseNameCtxKey).(string)

			listOptionsFunc := func(options *metav1.ListOptions) {
				options.LabelSelector = fmt.Sprintf("app.kubernetes.io/name=%s,app.kubernetes.io/instance=%s", defaultHelmChartName, helmReleaseName)
			}

			podList := corev1.PodList{}
			if err := config.Client().Resources().WithNamespace(defaultNamespace).List(ctx, &podList, listOptionsFunc); err != nil {
				subT.Fatal(err)
			}

			nodeNameToLineByLineMetricsMap := make(map[string][]string)
			k8sClient, _ := ctx.Value(k8sClientCtxKey).(*kubernetes.Clientset)

			portNumber := ctx.Value(portNumberCtxKey).(int)

			wg, lock := sync.WaitGroup{}, sync.Mutex{}
			for _, pod := range podList.Items {
				wg.Add(1)
				go func() {
					defer wg.Done()

					requiredMetricNameCheckMap := make(map[string]struct{})

					var requiredMetrics []string
					for trial := range iter.N(10) {
						res := k8sClient.CoreV1().RESTClient().Get().
							Namespace(pod.Namespace).
							Resource("pods").
							Name(fmt.Sprintf("%s:%d", pod.Name, portNumber)).
							SubResource("proxy").
							Suffix("metrics").
							Do(ctx)

						resBody, err := res.Raw()
						if err != nil {
							subT.Error(err)
							return
						}

						// furiosa_npu_alive, furiosa_npu_error,furiosa_npu_hw_power,furiosa_npu_hw_temperature
						lineByLineMetrics := strings.Split(string(resBody), "\n")
						requiredMetrics = make([]string, 0, len(lineByLineMetrics))
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
							// `NodeName` must be set because it is created by DaemonSet.
							lock.Lock()
							nodeNameToLineByLineMetricsMap[pod.Spec.NodeName] = requiredMetrics
							lock.Unlock()

							return
						}

						fmt.Printf("[Trial %d] Required metrics does not present on pod '%s'. Retry after 30 seconds.\n", trial, pod.Name)
						time.Sleep(time.Second * 30)
					}

					subT.Error(fmt.Errorf("unable to get required metrics from pod %s", pod.Name))
				}()
			}

			wg.Wait()

			ctx = context.WithValue(ctx, nodeNameToLineByLineMetricMapCtxKey, nodeNameToLineByLineMetricsMap)
			return ctx
		}).
		Assess("Validate collected furiosa metrics", func(ctx context.Context, subT *testing.T, config *envconf.Config) context.Context {
			nodeNameToLineByLineMetricsMap := ctx.Value(nodeNameToLineByLineMetricMapCtxKey).(map[string][]string)
			for nodeName, lineByLineMetrics := range nodeNameToLineByLineMetricsMap {
				checklistByDevice := make(map[string]map[string]map[string]struct{})
				for _, lineByLineMetric := range lineByLineMetrics {
					metricName, labels, err := parseMetricLineDataString(lineByLineMetric)
					if err != nil {
						subT.Fatal(err)
					}

					if labels["kubernetes_node_name"] != nodeName {
						subT.Fatal(fmt.Errorf("metric from node %s has a wrong label data 'kubernetes_node_name' with value %s", nodeName, labels["kubernetes_node_name"]))
					}

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

				for device, metricsWithLabelAttributesMap := range checklistByDevice {
					for metricName, labelAttributesMap := range metricsWithLabelAttributesMap {
						for _, requiredLabelAttribute := range requiredMetricLabelAttributes[metricName] {
							if _, ok := labelAttributesMap[requiredLabelAttribute]; !ok {
								subT.Errorf("required metric '%s' with label attribute '%s' does not present on node '%s' with device '%s'", metricName, requiredLabelAttribute, nodeName, device)
							}
						}
					}
				}
			}

			return ctx
		}).
		Feature()

	testenv.Test(t, f)
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
