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

Deploying Furiosa Metrics Exporter with Helm
---------------------------------------------------------
The Furiosa metrics exporter helm chart is available at https://github.com/furiosa-ai/helm-charts. To configure deployment as you need, you can modify ``charts/furiosa-metrics-exporter/values.yaml``.
For example, the Furiosa metrics exporter Helm chart automatically creates a Service Object with Prometheus annotations to enable metric scraping automatically. You can modify the values.yaml to change the port or disable the Prometheus annotations if needed.
You can deploy the Furiosa Metrics Exporter by running the following commands:

.. code-block:: sh

    helm repo add furiosa https://furiosa-ai.github.io/helm-charts
    helm repo update
    helm install furiosa-metrics-exporter furiosa/furiosa-metrics-exporter -n furiosa-system

Alternative Installation: furiosa-metrics-exporter via apt
----------------------------------------------------------
.. note::
   For special use cases or non-Kubernetes environments, you can alternatively install the exporter via an ``apt`` package.

The minimum requirements for installing the ``furiosa-metrics-exporter`` package are as follows:

* Ubuntu 22.04 LTS (Debian Bookworm) or later
* Linux Kernel 6.3 or later
* Administrator privileges on the system (root)
* Setting up AptSetup and InstallingPrerequisites

Then, please install the furiosa-metrics-exporter package as follows:

.. code-block:: sh

   sudo apt update
   sudo apt install -y furiosa-metrics-exporter

This command installs packages furiosa-libsmi and furiosa-metrics-exporter.

Configuration
-----------------------------------
The Furiosa Metrics Exporter has various command-line (CLI) options.
The following table summarizes the available CLI options:

.. list-table:: Furiosa Metrics Exporter CLI Options
   :align: center
   :widths: 150 80 350
   :header-rows: 1

   * - CLI Option
     - Required
     - Description
   * - ``--interval <int>``
     - Yes
     - Metrics collection interval in seconds. e.g., ``--interval 10`` for collecting metrics every 10 seconds.
   * - ``--port <int>``
     - Yes
     - Port number for the metrics HTTP server. e.g., ``--port 6254``.
   * - ``--node-name <string>``
     - No
     - If set, hostname label will be added to the metrics. Default is empty (``""``).
   * - ``--kube-resources-label <bool>``
     - No
     - If set, Kubernetes resource labels (such as namespace, pod, and container) will be added to the metrics. Default is ``false``. Recommended when running in Kubernetes for richer metrics context.


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
     - arch, core, device, uuid, pci_bus_id, firmware_version, driver_version, hostname, namespace, pod, container
     - The liveness of the Furiosa NPU device.
   * - Temperature
     - furiosa_npu_hw_temperature
     - gauge
     - arch, core, device, uuid, pci_bus_id, firmware_version, driver_version, hostname, namespace, pod, container, label
     - The temperature of the Furiosa NPU device.
   * - Power
     - furiosa_npu_hw_power
     - gauge
     - arch, core, device, uuid, pci_bus_id, firmware_version, driver_version, hostname, namespace, pod, container, label
     - The power consumption of the Furiosa NPU device.
   * - Core Utilization
     - furiosa_npu_core_utilization
     - gauge
     - arch, core, device, uuid, pci_bus_id, firmware_version, driver_version, hostname, namespace, pod, container
     - The core utilization of the Furiosa NPU device.
   * - Core Frequency
     - furiosa_npu_core_frequency
     - gauge
     - arch, core, device, uuid, pci_bus_id, firmware_version, driver_version, hostname, namespace, pod, container
     - The core frequency of the Furiosa NPU device.
   * - Cycle Count
     - furiosa_npu_total_cycle_count
     - counter
     - arch, core, device, uuid, pci_bus_id, firmware_version, driver_version, hostname, namespace, pod, container
     - The total cycle count of the Furiosa NPU device.
   * - Task Execution Cycle
     - furiosa_npu_task_execution_cycle
     - counter
     - arch, core, device, uuid, pci_bus_id, firmware_version, driver_version, hostname, namespace, pod, container
     - The task execution cycle of the NPU Task.
   * - Total DRAM Size
     - furiosa_npu_dram_total
     - gauge
     - arch, core, device, uuid, pci_bus_id, firmware_version, driver_version, hostname, namespace, pod, container
     - The total DRAM size of the Furiosa NPU device.
   * - DRAM Used Size
     - furiosa_npu_dram_usage
     - gauge
     - arch, core, device, uuid, pci_bus_id, firmware_version, driver_version, hostname, namespace, pod, container
     - The currently used DRAM size of the Furiosa NPU device.
   * - Throttling Events Count
     - furiosa_npu_throttling_events_count
     - gauge
     - arch, core, device, uuid, pci_bus_id, firmware_version, driver_version, hostname, namespace, pod, container
     - The number of throttling events that occurred on the Furiosa NPU device within a fixed time window.

All metrics share common metric labels such as arch, core, device, uuid, pci_bus_id, firmware_version, driver_version, hostname, namespace, pod, and container.
The following table describes the common metric labels:

.. list-table:: Common NPU Metrics Label Attributes
   :align: center
   :widths: 100 300
   :header-rows: 1

   * - Attribute Name
     - Description
   * - arch
     - The architecture of the Furiosa NPU device. e.g. warboy, rngd
   * - core
     - The core number of the Furiosa NPU device. e.g. 0, 1, 2, 3, 4, 5, 6, 7, 0-7
   * - device
     - The device name of the Furiosa NPU device. e.g. npu0
   * - uuid
     - The UUID of the Furiosa NPU device.
   * - pci_bus_id
     - The PCI bus ID of the Furiosa NPU device. e.g. 0000:c7:00.0
   * - firmware_version
     - The firmware version of the Furiosa NPU device. e.g. 2025.1.0+696efad
   * - driver_version
     - The driver version of the Furiosa NPU device. e.g. 2025.1.0+f09a8d8


The following metric labels are optional and enabled only when the corresponding command-line options (``--node-name`` and ``--kube-resources-label``) are set.
Additionally, the *namespace*, *pod*, and *container* labels require an environment where the `Kubernetes PodResource API <https://kubernetes.io/blog/2023/08/23/kubelet-podresources-api-ga/>`_ is supported.

.. list-table:: Optional NPU Metrics Label Attributes
   :align: center
   :widths: 100 300
   :header-rows: 1

   * - Attribute Name
     - Description
   * - hostname
     - The hostname of the machine where the exporter is running.
   * - namespace
     - The namespace of the Kubernetes pod to which the NPU is allocated.
   * - pod
     - The name of the Kubernetes pod to which the NPU is allocated.
   * - container
     - The container name within the Kubernetes pod to which the NPU is allocated.

The metric label “label” is used to describe additional attributes specific to each metric.
This approach helps avoid having too many metric definitions and effectively aggregates metrics that share common characteristics.

.. list-table:: Additional Metric Label Attributes
   :align: center
   :widths: 100 120 200
   :header-rows: 1

   * - Metric Type
     - Label Name
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


**Examples**

The following shows real-world example of the metrics:

.. code-block:: sh

  #liveness
  furiosa_npu_alive{arch="rngd",container="furiosa",core="0-7",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 1

  #temperature
  furiosa_npu_hw_temperature{arch="rngd",container="furiosa",core="0-7",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",label="ambient",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 52
  furiosa_npu_hw_temperature{arch="rngd",container="furiosa",core="0-7",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",label="peak",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 67.756

  #power
  furiosa_npu_hw_power{arch="rngd",container="furiosa",core="0-7",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",label="rms",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 50

  #core utilization
  furiosa_npu_core_utilization{arch="rngd",container="furiosa",core="0",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 99.68363645361265
  furiosa_npu_core_utilization{arch="rngd",container="furiosa",core="1",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 99.68363645361265
  furiosa_npu_core_utilization{arch="rngd",container="furiosa",core="2",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 99.68363645361265
  furiosa_npu_core_utilization{arch="rngd",container="furiosa",core="3",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 99.68363645361265
  furiosa_npu_core_utilization{arch="rngd",container="furiosa",core="4",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 99.6826341187199
  furiosa_npu_core_utilization{arch="rngd",container="furiosa",core="5",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 99.6826341187199
  furiosa_npu_core_utilization{arch="rngd",container="furiosa",core="6",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 99.6826341187199
  furiosa_npu_core_utilization{arch="rngd",container="furiosa",core="7",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 99.6826341187199

  #core frequency
  furiosa_npu_core_frequency{arch="rngd",container="furiosa",core="0",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 1750
  furiosa_npu_core_frequency{arch="rngd",container="furiosa",core="1",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 1750
  furiosa_npu_core_frequency{arch="rngd",container="furiosa",core="2",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 1750
  furiosa_npu_core_frequency{arch="rngd",container="furiosa",core="3",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 1750
  furiosa_npu_core_frequency{arch="rngd",container="furiosa",core="4",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 1750
  furiosa_npu_core_frequency{arch="rngd",container="furiosa",core="5",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 1750
  furiosa_npu_core_frequency{arch="rngd",container="furiosa",core="6",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 1750
  furiosa_npu_core_frequency{arch="rngd",container="furiosa",core="7",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 1750

  #total cycle count
  furiosa_npu_total_cycle_count{arch="rngd",container="furiosa",core="0",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 1.7242541456e+10
  furiosa_npu_total_cycle_count{arch="rngd",container="furiosa",core="1",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 1.7242541456e+10
  furiosa_npu_total_cycle_count{arch="rngd",container="furiosa",core="2",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 1.7242541456e+10
  furiosa_npu_total_cycle_count{arch="rngd",container="furiosa",core="3",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 1.7242541456e+10
  furiosa_npu_total_cycle_count{arch="rngd",container="furiosa",core="4",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 1.7175902913e+10
  furiosa_npu_total_cycle_count{arch="rngd",container="furiosa",core="5",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 1.7175902913e+10
  furiosa_npu_total_cycle_count{arch="rngd",container="furiosa",core="6",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 1.7175902913e+10
  furiosa_npu_total_cycle_count{arch="rngd",container="furiosa",core="7",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 1.7175902913e+10

  #task execution cycle
  furiosa_npu_task_execution_cycle{arch="rngd",container="furiosa",core="0",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 5.686392711e+09
  furiosa_npu_task_execution_cycle{arch="rngd",container="furiosa",core="1",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 5.686392711e+09
  furiosa_npu_task_execution_cycle{arch="rngd",container="furiosa",core="2",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 5.686392711e+09
  furiosa_npu_task_execution_cycle{arch="rngd",container="furiosa",core="3",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 5.686392711e+09
  furiosa_npu_task_execution_cycle{arch="rngd",container="furiosa",core="4",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 5.685170235e+09
  furiosa_npu_task_execution_cycle{arch="rngd",container="furiosa",core="5",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 5.685170235e+09
  furiosa_npu_task_execution_cycle{arch="rngd",container="furiosa",core="6",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 5.685170235e+09
  furiosa_npu_task_execution_cycle{arch="rngd",container="furiosa",core="7",device="npu0",driver_version="2025.1.0+f09a8d8",firmware_version="2025.1.0+696efad",hostname="cntk002",namespace="default",pci_bus_id="0000:c7:00.0",pod="furiosa",uuid="09512C86-0702-4303-8F40-474746474A40"} 5.685170235e+09



License
-------

.. code-block:: text

   Copyright 2023 FuriosaAI, Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.