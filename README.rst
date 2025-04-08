.. _MetricsExporter:

###################################
Installing Furiosa Metrics Exporter
###################################


Furiosa Metrics Exporter
================================================================
The Furiosa metrics exporter exposes collection of metrics related to
FuriosaAI NPU devices in `Prometheus <https://prometheus.io/>`_ format.
In a Kubernetes cluster, you can scrape the metrics provided by furiosa-metrics-exporter
using Prometheus and visualize them with a Grafana dashboard.
This can be easily set up using the `Prometheus Chart <https://github.com/prometheus-community/helm-charts/tree/main/charts/prometheus>`_
and `Grafana <https://github.com/grafana/helm-charts/tree/main/charts/grafana>`_
Helm charts, along with the furiosa-metrics-exporter Helm chart.


Metrics
-----------------------------------
The exporter is composed of chain of collectors, each collector is responsible
for collecting specific metrics from the Furiosa NPU devices.
The following table shows the available collectors and metrics:


.. list-table:: NPU Metrics
   :align: center
   :widths: 100 100 100 100 200
   :header-rows: 1

   * - Collector Name
     - Metric
     - Type
     - Metric Labels
     - Description
   * - Liveness
     - furiosa_npu_alive
     - gauge
     - arch, core, device, uuid, kubernetes_node_name
     - The liveness of the Furiosa NPU device.
   * - Temperature
     - furiosa_npu_hw_temperature
     - gauge
     - arch, core, device, uuid, kubernetes_node_name, label
     - The temperature of the Furiosa NPU device.
   * - Power
     - furiosa_npu_hw_power
     - gauge
     - arch, core, device, uuid, kubernetes_node_name, label
     - The power consumption of the Furiosa NPU device.
   * - Core Utilization
     - furiosa_npu_core_utilization
     - gauge
     - arch, core, device, uuid, kubernetes_node_name
     - The core utilization of the Furiosa NPU device.

All metrics share common metric labels such as arch, core, device, kubernetes_node_name, and uuid.
The following table describes the common metric labels:

.. list-table:: Common NPU Metrics Label
   :align: center
   :widths: 100 300
   :header-rows: 1

   * - Common Metric Label
     - Description
   * - arch
     - The architecture of the Furiosa NPU device. e.g. warboy, rngd
   * - core
     - The core number of the Furiosa NPU device. e.g. 0, 1, 2, 3, 4, 5, 6, 7, 0-1, 2-3, 0-3, 4-5, 6-7, 4-7, 0-7
   * - device
     - The device name of the Furiosa NPU device. e.g. npu0
   * - kubernetes_node_name
     - The name of the Kubernetes node where the exporter is running, this attribute can be missing if the exporter is running on the host machine or in a naked container.
   * - uuid
     - The UUID of the Furiosa NPU device.

The metric label “label” is used to describe additional attributes specific to each metric.
This approach helps avoid having too many metric definitions and effectively aggregates metrics that share common characteristics.

.. list-table:: NPU Metrics Type
   :align: center
   :widths: 100 120 200
   :header-rows: 1

   * - Metric Type
     - Label Attribute
     - Description
   * - Temperature
     - peak
     - The highest temperature observed from SoC sensors
   * - Temperature
     - ambient
     - The temperature observed from sensors attached to the board
   * - Power
     - rms
     - Root Mean Square (RMS) value of the power consumed by the device, providing an average power consumption metric over a period of time.


The following shows real-world example of the metrics:

.. code-block:: sh

  #liveness
  furiosa_npu_alive{arch="rngd",core="0-7",device="npu0",kubernetes_node_name="node",uuid="uuid"} 1

  #temperature
  furiosa_npu_hw_temperature{arch="rngd",core="0-7",device="npu0",kubernetes_node_name="node",label="peak",uuid="uuid"} 39
  furiosa_npu_hw_temperature{arch="rngd",core="0-7",device="npu0",kubernetes_node_name="node",label="ambient",uuid="uuid"} 35

  #power
  furiosa_npu_hw_power{arch="rngd",core="0-7",device="npu0",kubernetes_node_name="node",label="rms",uuid="uuid"} 4795000

  #core utilization
  furiosa_npu_core_utilization{arch="rngd",core="0",device="npu0",kubernetes_node_name="node",uuid="uuid"} 90
  furiosa_npu_core_utilization{arch="rngd",core="1",device="npu0",kubernetes_node_name="node",uuid="uuid"} 90
  furiosa_npu_core_utilization{arch="rngd",core="2",device="npu0",kubernetes_node_name="node",uuid="uuid"} 90
  furiosa_npu_core_utilization{arch="rngd",core="3",device="npu0",kubernetes_node_name="node",uuid="uuid"} 90
  furiosa_npu_core_utilization{arch="rngd",core="4",device="npu0",kubernetes_node_name="node",uuid="uuid"} 90
  furiosa_npu_core_utilization{arch="rngd",core="5",device="npu0",kubernetes_node_name="node",uuid="uuid"} 90
  furiosa_npu_core_utilization{arch="rngd",core="6",device="npu0",kubernetes_node_name="node",uuid="uuid"} 90
  furiosa_npu_core_utilization{arch="rngd",core="7",device="npu0",kubernetes_node_name="node",uuid="uuid"} 90

Deploying Furiosa Metrics Exporter with Helm
---------------------------------------------------------
The Furiosa metrics exporter helm chart is available at https://github.com/furiosa-ai/helm-charts. To configure deployment as you need, you can modify ``charts/furiosa-metrics-exporter/values.yaml``.
For example, the Furiosa metrics exporter Helm chart automatically creates a Service Object with Prometheus annotations to enable metric scraping automatically. You can modify the values.yaml to change the port or disable the Prometheus annotations if needed.
You can deploy the Furiosa Metrics Exporter by running the following commands:

.. code-block:: sh

    helm repo add furiosa https://furiosa-ai.github.io/helm-charts
    helm repo update
    helm install furiosa-metrics-exporter furiosa/furiosa-metrics-exporter -n kube-system

