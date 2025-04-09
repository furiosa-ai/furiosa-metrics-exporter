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
     - arch, core, device, uuid, pci_bud_id, firmware_version, pert_version, driver_version, hostname, namespace, pod, container
     - The liveness of the Furiosa NPU device.
   * - Temperature
     - furiosa_npu_hw_temperature
     - gauge
     - arch, core, device, uuid, pci_bud_id, firmware_version, pert_version, driver_version, hostname, namespace, pod, container, label
     - The temperature of the Furiosa NPU device.
   * - Power
     - furiosa_npu_hw_power
     - gauge
     - arch, core, device, uuid, pci_bud_id, firmware_version, pert_version, driver_version, hostname, namespace, pod, container, label
     - The power consumption of the Furiosa NPU device.
   * - Core Utilization
     - furiosa_npu_core_utilization
     - gauge
     - arch, core, device, uuid, pci_bud_id, firmware_version, pert_version, driver_version, hostname, namespace, pod, container
     - The core utilization of the Furiosa NPU device.
   * - Core Frequency
     - furiosa_npu_core_frequency
     - gauge
     - arch, core, device, uuid, pci_bud_id, firmware_version, pert_version, driver_version, hostname, namespace, pod, container
     - The core utilization of the Furiosa NPU device.
   * - Cycle Count
     - furiosa_npu_total_cycle_count
     - counter
     - arch, core, device, uuid, pci_bud_id, firmware_version, pert_version, driver_version, hostname, namespace, pod, container
     - The total cycle count of the Furiosa NPU device.
   * - Task Execution Cycle
     - furiosa_npu_task_execution_cycle
     - counter
     - arch, core, device, uuid, pci_bud_id, firmware_version, pert_version, driver_version, hostname, namespace, pod, container
     - The task execution cycle of the NPU Task.

All metrics share common metric labels such as arch, core, device, uuid, pci_bud_id, firmware_version, pert_version, driver_version, hostname, namespace, pod, and container.
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
   * - uuid
     - The UUID of the Furiosa NPU device.
   * - pci_bus_id
     - The PCI bus ID of the Furiosa NPU device. e.g. 0000:c7:00.0
   * - firmware_version
     - The firmware version of the Furiosa NPU device. e.g. 2025.1.0+696efad
   * - pert_version
     - The pert version of the Furiosa NPU device. e.g. 2025.1.0+1694e18
   * - driver_version
     - The driver version of the Furiosa NPU device. e.g. 2025.1.0+f09a8d8
   * - hostname
     - The hostname of the machine where the exporter is running. This attribute can be missing if the exporter is running on the host machine or in a naked container.
   * - namespace
     - The Kubernetes namespace where the exporter is running. This attribute can be missing if the exporter is running on the host machine or in a naked container.
   * - pod
     - The name of the Kubernetes pod where the exporter is running. This attribute can be missing if the exporter is running on the host machine or in a naked container.
   * - container
     - The name of the Kubernetes container where the exporter is running. This attribute can be missing if the exporter is running on the host machine or in a naked container.

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
  furiosa_npu_alive{arch="rngd",container="furiosa",core="0-7",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 1

  #temperature
  furiosa_npu_hw_temperature{arch="rngd",container="furiosa",core="0-7",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",label="ambient",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 52
  furiosa_npu_hw_temperature{arch="rngd",container="furiosa",core="0-7",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",label="peak",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 67.756

  #power
  furiosa_npu_hw_power{arch="rngd",container="furiosa",core="0-7",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",label="rms",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 50

  #core utilization
  furiosa_npu_core_utilization{arch="rngd",container="furiosa",core="0",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 99.68363645361265
  furiosa_npu_core_utilization{arch="rngd",container="furiosa",core="1",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 99.68363645361265
  furiosa_npu_core_utilization{arch="rngd",container="furiosa",core="2",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 99.68363645361265
  furiosa_npu_core_utilization{arch="rngd",container="furiosa",core="3",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 99.68363645361265
  furiosa_npu_core_utilization{arch="rngd",container="furiosa",core="4",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 99.6826341187199
  furiosa_npu_core_utilization{arch="rngd",container="furiosa",core="5",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 99.6826341187199
  furiosa_npu_core_utilization{arch="rngd",container="furiosa",core="6",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 99.6826341187199
  furiosa_npu_core_utilization{arch="rngd",container="furiosa",core="7",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 99.6826341187199

  #core frequency
  furiosa_npu_core_frequency{arch="rngd",container="furiosa",core="0",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 1750
  furiosa_npu_core_frequency{arch="rngd",container="furiosa",core="1",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 1750
  furiosa_npu_core_frequency{arch="rngd",container="furiosa",core="2",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 1750
  furiosa_npu_core_frequency{arch="rngd",container="furiosa",core="3",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 1750
  furiosa_npu_core_frequency{arch="rngd",container="furiosa",core="4",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 1750
  furiosa_npu_core_frequency{arch="rngd",container="furiosa",core="5",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 1750
  furiosa_npu_core_frequency{arch="rngd",container="furiosa",core="6",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 1750
  furiosa_npu_core_frequency{arch="rngd",container="furiosa",core="7",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 1750

  #total cycle count
  furiosa_npu_total_cycle_count{arch="rngd",container="furiosa",core="0",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 1.7242541456e+10
  furiosa_npu_total_cycle_count{arch="rngd",container="furiosa",core="1",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 1.7242541456e+10
  furiosa_npu_total_cycle_count{arch="rngd",container="furiosa",core="2",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 1.7242541456e+10
  furiosa_npu_total_cycle_count{arch="rngd",container="furiosa",core="3",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 1.7242541456e+10
  furiosa_npu_total_cycle_count{arch="rngd",container="furiosa",core="4",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 1.7175902913e+10
  furiosa_npu_total_cycle_count{arch="rngd",container="furiosa",core="5",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 1.7175902913e+10
  furiosa_npu_total_cycle_count{arch="rngd",container="furiosa",core="6",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 1.7175902913e+10
  furiosa_npu_total_cycle_count{arch="rngd",container="furiosa",core="7",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 1.7175902913e+10

  #task execution cycle
  furiosa_npu_task_execution_cycle{arch="rngd",container="furiosa",core="0",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 5.686392711e+09
  furiosa_npu_task_execution_cycle{arch="rngd",container="furiosa",core="1",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 5.686392711e+09
  furiosa_npu_task_execution_cycle{arch="rngd",container="furiosa",core="2",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 5.686392711e+09
  furiosa_npu_task_execution_cycle{arch="rngd",container="furiosa",core="3",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 5.686392711e+09
  furiosa_npu_task_execution_cycle{arch="rngd",container="furiosa",core="4",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 5.685170235e+09
  furiosa_npu_task_execution_cycle{arch="rngd",container="furiosa",core="5",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 5.685170235e+09
  furiosa_npu_task_execution_cycle{arch="rngd",container="furiosa",core="6",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 5.685170235e+09
  furiosa_npu_task_execution_cycle{arch="rngd",container="furiosa",core="7",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pert_version="2025.1.0+1694e18",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 5.685170235e+09



Deploying Furiosa Metrics Exporter with Helm
---------------------------------------------------------
The Furiosa metrics exporter helm chart is available at https://github.com/furiosa-ai/helm-charts.

To configure deployment as you need, you can modify ``charts/furiosa-metrics-exporter/values.yaml``.
For example, the Furiosa metrics exporter Helm chart automatically creates a Service Object with Prometheus annotations to enable metric scraping automatically.
You can modify the values.yaml to change the port or disable the Prometheus annotations if needed.

You can deploy the Furiosa Metrics Exporter by running the following commands:

.. code-block:: sh

    helm repo add furiosa https://furiosa-ai.github.io/helm-charts
    helm repo update
    helm install furiosa-metrics-exporter furiosa/furiosa-metrics-exporter -n kube-system

