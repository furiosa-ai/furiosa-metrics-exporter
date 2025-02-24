package kubernetes

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"

	podResourcesAPI "k8s.io/kubelet/pkg/apis/podresources/v1alpha1"
)

const (
	k8sSocket                         = "/var/lib/kubelet/pod-resources/kubelet.sock"
	furiosaResourcePrefix             = "furiosa.ai"
	furiosaPartitionedResourcePattern = "_cores_"
)

type PodInfo struct {
	Name      string
	Namespace string
}

func ConnectToServer() (*grpc.ClientConn, func(), error) {
	resolver.SetDefaultScheme("passthrough")

	conn, err := grpc.NewClient(k8sSocket,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, "unix", addr)
		}))

	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to connect to '%s'; err: %w", k8sSocket, err)
	}

	return conn, func() { conn.Close() }, nil
}

func ListPods(conn *grpc.ClientConn) (*podResourcesAPI.ListPodResourcesResponse, error) {
	client := podResourcesAPI.NewPodResourcesListerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.List(ctx, &podResourcesAPI.ListPodResourcesRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod resources; err: %w", err)
	}

	return resp, nil
}

func GenerateDeviceMap(devicePods *podResourcesAPI.ListPodResourcesResponse) map[string]PodInfo {
	deviceMap := make(map[string]PodInfo)

	for _, pod := range devicePods.GetPodResources() {
		for _, container := range pod.GetContainers() {
			for _, device := range container.GetDevices() {

				resource := device.GetResourceName()
				if !strings.HasPrefix(resource, furiosaResourcePrefix) { // Need to check device type? (e.g., warboy, rngd)
					continue
				}

				podName := pod.GetName()
				podNamespace := pod.GetNamespace()

				podInfo := PodInfo{
					Name:      podName,
					Namespace: podNamespace,
				}

				for _, deviceID := range device.GetDeviceIds() {
					// TODO(jongchan): partitioned device case

					deviceMap[deviceID] = podInfo
				}
			}
		}
	}
	return deviceMap

}

func ContainsPECore(deviceID string, core string) bool {
	if !strings.Contains(deviceID, furiosaPartitionedResourcePattern) {
		return true
	} else {
		cores := strings.Split(deviceID, furiosaPartitionedResourcePattern)[1]
		coreRange := strings.Split(cores, "-")
		if len(coreRange) == 1 {
			return coreRange[0] == core
		} else {
			n, _ := strconv.Atoi(coreRange[0])
			m, _ := strconv.Atoi(coreRange[1])
			coreNum, _ := strconv.Atoi(core)
			return coreNum >= n && coreNum <= m
		}
	}
}

func GetCoreNum(deviceID string) string {
	if !strings.Contains(deviceID, furiosaPartitionedResourcePattern) {
		return "0-7"
	} else {
		return strings.Split(deviceID, furiosaPartitionedResourcePattern)[1]
	}
}

func IsPartionedDevice(deviceID string) bool {
	return strings.Contains(deviceID, furiosaPartitionedResourcePattern)
}
