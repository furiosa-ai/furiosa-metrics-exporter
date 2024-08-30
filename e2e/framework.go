package e2e

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/gomega"
	"k8s.io/client-go/tools/clientcmd"

	helmclient "github.com/mittwald/go-helm-client"
	"k8s.io/client-go/kubernetes"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	defaultKubeConfigPath = ".kube/config"
	defaultNS             = "kube-system"

	kubeConfigEnvKey = "KUBECONFIG"
)

type Framework interface {
	GetClientSet() clientset.Interface
	GetNamespace() string

	DeployHelmChart(releaseName string, chartPath string, values string) func()
	DeleteHelmChart() func()
}

// framework is container for components can be reused for each test
type framework struct {
	ClientConfig *rest.Config
	ClientSet    clientset.Interface

	Namespace string

	HelmClient helmclient.Client
	HelmChart  *helmclient.ChartSpec
}

func NewFrameworkWithDefaultNamespace() (Framework, error) {
	return NewFrameworkWithNamespace(defaultNS)
}

func NewFrameworkWithNamespace(namespace string) (Framework, error) {
	kubeConfigPath := os.Getenv(kubeConfigEnvKey)
	if kubeConfigPath == "" {
		fmt.Println("KUBECONFIG environment variable not set. Use default path.")

		homePath, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}

		kubeConfigPath = filepath.Join(homePath, defaultKubeConfigPath)
	}

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, err
	}

	clientSet, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}

	helmChartClient, err := helmclient.NewClientFromRestConf(
		&helmclient.RestConfClientOptions{
			Options: &helmclient.Options{
				Namespace: namespace,
			},
			RestConfig: kubeConfig,
		},
	)
	if err != nil {
		return nil, err
	}

	return &framework{
		ClientConfig: kubeConfig,
		ClientSet:    clientSet,
		Namespace:    namespace,
		HelmClient:   helmChartClient,
		HelmChart:    nil,
	}, nil
}

func (frk *framework) GetClientSet() clientset.Interface {
	return frk.ClientSet
}

func (frk *framework) GetNamespace() string {
	return frk.Namespace
}

func (frk *framework) DeployHelmChart(releaseName string, chartPath string, values string) func() {
	return func() {
		helmChartSpec := &helmclient.ChartSpec{
			ReleaseName:     fmt.Sprintf("%s-%s", releaseName, generateRandomString()),
			ChartName:       chartPath,
			Namespace:       frk.Namespace,
			CreateNamespace: false,
			Wait:            false,
			Timeout:         5 * time.Minute,
			CleanupOnFail:   false,
			ValuesYaml:      values,
		}
		frk.HelmChart = helmChartSpec

		_, err := frk.HelmClient.InstallChart(context.TODO(), frk.HelmChart, nil)
		Expect(err).To(BeNil())
	}
}

func (frk *framework) DeleteHelmChart() func() {
	return func() {
		err := frk.HelmClient.UninstallRelease(frk.HelmChart)
		Expect(err).To(BeNil())
	}
}

func DeployTest(frk Framework, releaseName string, chartPath string, values string) func() {
	return func() {
		helmChartSpec := &helmclient.ChartSpec{
			ReleaseName:     fmt.Sprintf("%s-%s", releaseName, generateRandomString()),
			ChartName:       chartPath,
			Namespace:       frk.(*framework).Namespace,
			CreateNamespace: false,
			Wait:            false,
			Timeout:         5 * time.Minute,
			CleanupOnFail:   false,
			ValuesYaml:      values,
		}
		frk.(*framework).HelmChart = helmChartSpec

		_, err := frk.(*framework).HelmClient.InstallChart(context.TODO(), frk.(*framework).HelmChart, nil)
		Expect(err).To(BeNil())
	}
}
