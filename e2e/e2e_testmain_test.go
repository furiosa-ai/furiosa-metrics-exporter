package e2e_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	helmclient "github.com/mittwald/go-helm-client"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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

	defaultHelmChartName = "furiosa-metrics-exporter"

	helmClientCtxKey = "HELM_CLIENT"
)

var (
	testenv env.Environment = nil
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
		// Setup Helm Client
		func(ctx context.Context, config *envconf.Config) (context.Context, error) {
			restConfig, err := clientcmd.BuildConfigFromFlags("", config.KubeconfigFile())
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

			ctx = context.WithValue(ctx, helmClientCtxKey, helmClient)

			return ctx, nil
		},
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

			return ctx, nil
		},
	)

	testenv.Finish(
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

func TestSample(t *testing.T) {
	f := features.New("TEST").
		Assess("Check `Service` is created", func(ctx context.Context, subT *testing.T, config *envconf.Config) context.Context {
			service := corev1.Service{}
			if err := config.Client().Resources().Get(ctx, defaultHelmChartName, defaultNamespace, &service); err != nil {
				subT.Error(err)
			}

			return ctx
		}).
		Assess("Wait until `DaemonSet` is ready", func(ctx context.Context, subT *testing.T, config *envconf.Config) context.Context {
			daemonset := appsv1.DaemonSet{}
			if err := config.Client().Resources().Get(ctx, defaultHelmChartName, defaultNamespace, &daemonset); err != nil {
				subT.Error(err)
			}

			waitCondition := conditions.New(config.Client().Resources()).DaemonSetReady(&daemonset)
			waitTimeout := wait.WithTimeout(time.Minute * 1)
			waitInterval := wait.WithInterval(time.Second * 10)

			if err := wait.For(waitCondition, waitTimeout, waitInterval); err != nil {
				subT.Error(err)
			}

			return ctx
		}).
		Feature()

	testenv.Test(t, f)
}
