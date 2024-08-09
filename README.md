# Furiosa Metrics Exporter 

## Overview
This repository contains Furiosa Metric Exporter implementation, and it exposes collection of metrics related to FuriosaAI NPU devices in [Prometheus](https://prometheus.io/) format.

## Metrics
The exporter is composed of chain of collectors, each collector is responsible for collecting specific metrics from the Furiosa NPU devices.
The following table shows the available collectors and metrics:

| Collector Name | Metric                     | Type  | Metric Labels                                         | Description                                      |
|----------------|----------------------------|-------|-------------------------------------------------------|--------------------------------------------------|
| Liveness       | furiosa_npu_alive          | gauge | arch, core, device, uuid, kubernetes_node_name        | The liveness of the Furiosa NPU device.          |
| Error          | furiosa_npu_error          | gauge | arch, core, device, uuid, kubernetes_node_name, label | The error count of the Furiosa NPU device.       |
| Temperature    | furiosa_npu_hw_temperature | gauge | arch, core, device, uuid, kubernetes_node_name, label | The temperature of the Furiosa NPU device.       |
| Power          | furiosa_npu_hw_power       | gauge | arch, core, device, uuid, kubernetes_node_name, label | The power consumption of the Furiosa NPU device. |

All metrics share common metric labels such as arch, core, device, kubernetes_node_name, and uuid.
The following table describes the common metric labels:

| Common Metric Label  | Description                                                                                                                                                          |
|----------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| arch                 | The architecture of the Furiosa NPU device. e.g. warboy, rngd                                                                                                        |
| core                 | The core number of the Furiosa NPU device. e.g. 0, 1, 2, 3, 4, 5, 6, 7, 0-1, 2-3, 0-3, 4-5, 6-7, 4-7, 0-7                                                            |
| device               | The device name of the Furiosa NPU device. e.g. npu0                                                                                                                 |
| kubernetes_node_name | The name of the Kubernetes node where the exporter is running, this attribute can be missing if the exporter is running on the host machine or in a naked container. |
| uuid                 | The UUID of the Furiosa NPU device.                                                                                                                                  |

The metric label “label” is used to describe additional attributes specific to each metric.
This approach helps avoid having too many metric definitions and effectively aggregates metrics that share common characteristics.

| Metric Type | Label Attribute    | Description                                                                                                                            |
|-------------|--------------------|----------------------------------------------------------------------------------------------------------------------------------------|
| Error       | axi_post_error     | Indicates count of axi post error.                                                                                                     |
|             | axi_fetch_error    | Indicates count of axi fetch error.                                                                                                    |
|             | axi_discard_error  | Indicates count of axi discard error.                                                                                                  |
|             | axi_doorbell_done  | Indicates count of axi doorbell done error.                                                                                            |
|             | pcie_post_error    | Indicates count of PCIe post error.                                                                                                    |
|             | pcie_fetch_error   | Indicates count of PCIe fetch error.                                                                                                   |
|             | pcie_discard_error | Indicates count of PCIe discard error.                                                                                                 |
|             | pcie_doorbell_done | Indicates count of PCIe doorbell done error.                                                                                           |
|             | device_error       | Total count of device error.                                                                                                           |
| Temperature | peak               | The highest temperature observed from SoC sensors                                                                                      |
|             | ambient            | The temperature observed from sensors attached to the board                                                                            |
| Power       | rms                | Root Mean Square (RMS) value of the power consumed by the device, providing an average power consumption metric over a period of time. |



The following shows real-world example of the metrics:
```shell
#liveness
furiosa_npu_alive{arch="rngd",core="0-7",device="npu0",kubernetes_node_name="node",uuid="uuid"} 1

#error
furiosa_npu_error{arch="rngd",core="0-7",device="npu0",kubernetes_node_name="node",label="axi_post_error",uuid="uuid"} 0
furiosa_npu_error{arch="rngd",core="0-7",device="npu0",kubernetes_node_name="node",label="axi_fetch_error",uuid="uuid"} 0
furiosa_npu_error{arch="rngd",core="0-7",device="npu0",kubernetes_node_name="node",label="axi_discard_error",uuid="uuid"} 0
furiosa_npu_error{arch="rngd",core="0-7",device="npu0",kubernetes_node_name="node",label="axi_doorbell_done",uuid="uuid"} 0
furiosa_npu_error{arch="rngd",core="0-7",device="npu0",kubernetes_node_name="node",label="pcie_post_error",uuid="uuid"} 0
furiosa_npu_error{arch="rngd",core="0-7",device="npu0",kubernetes_node_name="node",label="pcie_fetch_error",uuid="uuid"} 0
furiosa_npu_error{arch="rngd",core="0-7",device="npu0",kubernetes_node_name="node",label="pcie_discard_error",uuid="uuid"} 0
furiosa_npu_error{arch="rngd",core="0-7",device="npu0",kubernetes_node_name="node",label="pcie_doorbell_done",uuid="uuid"} 0
furiosa_npu_error{arch="rngd",core="0-7",device="npu0",kubernetes_node_name="node",label="device_error",uuid="uuid"} 0

#temperature
furiosa_npu_hw_temperature{arch="rngd",core="0-7",device="npu0",kubernetes_node_name="node",label="peak",uuid="uuid"} 39
furiosa_npu_hw_temperature{arch="rngd",core="0-7",device="npu0",kubernetes_node_name="node",label="ambient",uuid="uuid"} 35

#power
furiosa_npu_hw_power{arch="rngd",core="0-7",device="npu0",kubernetes_node_name="node",label="rms",uuid="uuid"} 4795000
```

## Deployment

<!-- add baremetal support here -->

### Kubernetes
The helm chart is available at [deployments/helm](deployments/helm) directory. To configure deployment as you need, you can modify [deployments/helm/values.yaml](deployments/helm/values.yaml).
<!-- add prometheus annotation info here -->
<!-- add grafana dashboard import here -->
