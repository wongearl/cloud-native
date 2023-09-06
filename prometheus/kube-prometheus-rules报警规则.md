## Kube-prometheus rules 报警规则

## 1. alertmanager



```yaml
groups:
- name: alertmanager.rules
  rules:
  - alert: Alertmanager配置不一致
    annotations:
      message: '{{ $labels.cluster }} 集群Alertmanager集群的节点之间配置不同步 {{ $labels.service }}！'
    expr: |
      count_values("config_hash", alertmanager_config_hash{job="alertmanager-main",namespace="monitoring"}) BY (cluster, service) / ON(cluster, service) GROUP_LEFT() label_replace(max(prometheus_operator_spec_replicas{job="prometheus-operator",namespace="monitoring",controller="alertmanager"}) by (cluster, name, job, namespace, controller), "service", "alertmanager-$1", "name", "(.*)") != 1
    for: 5m
    labels:
      severity: critical
  - alert: Alertmanager重载失败
    annotations:
      message: '{{ $labels.cluster }} 集群在重载Alertmanager配置时失败 {{ $labels.namespace }}/{{ $labels.pod }}！'
    expr: |
      alertmanager_config_last_reload_successful{job="alertmanager-main",namespace="monitoring"} == 0
    for: 10m
    labels:
      severity: warning
  - alert: Alertmanager成员不一致
    annotations:
      message: '{{ $labels.cluster }} 集群Alertmanager未找到群集的所有成员！'
    expr: |
      alertmanager_cluster_members{job="alertmanager-main",namespace="monitoring"}
        != on (cluster,service) GROUP_LEFT()
      count by (cluster,service) (alertmanager_cluster_members{job="alertmanager-main",namespace="monitoring"})
    for: 5m
    labels:
      severity: critical
```



## 2. apiserver



```yaml
groups:
- name: kubernetes-system-apiserver
  rules:
  - alert: K8S的APISERVER潜在危险过高
    annotations:
      message: '{{ $labels.cluster }} 集群 API server 的 {{ $labels.verb }} {{ $labels.resource }} 有异常延迟 {{ $value }} 秒！'
    expr: |
      (
        cluster:apiserver_request_duration_seconds:mean5m{job="apiserver"}
        >
        on (verb) group_left()
        (
          avg by (verb) (cluster:apiserver_request_duration_seconds:mean5m{job="apiserver"} >= 0)
          +
          2*stddev by (verb) (cluster:apiserver_request_duration_seconds:mean5m{job="apiserver"} >= 0)
        )
      ) > on (verb) group_left()
      1.2 * avg by (verb) (cluster:apiserver_request_duration_seconds:mean5m{job="apiserver"} >= 0)
      and on (verb,resource)
      cluster_quantile:apiserver_request_duration_seconds:histogram_quantile{job="apiserver",quantile="0.99"}
      >
      1
    for: 5m
    labels:
      severity: warning
  - alert: K8S的APISERVER潜在致命风险
    annotations:
      message: '{{ $labels.cluster }} 集群 API server 的 {{ $labels.verb }} {{ $labels.resource }} 有 99% 的请求的延迟达 {{ $value }} 秒！'
    expr: |
      cluster_quantile:apiserver_request_duration_seconds:histogram_quantile{job="apiserver",quantile="0.99"} > 4
    for: 10m
    labels:
      severity: critical
  - alert: K8S的APISERVER存在返回错误过高
    annotations:
      message: '{{ $labels.cluster }} 集群 API server 请求中有 {{ $value | humanizePercentage }} 的返回错误！'
    expr: |
      sum(rate(apiserver_request_total{job="apiserver",code=~"5.."}[5m]))
        /
      sum(rate(apiserver_request_total{job="apiserver"}[5m])) > 0.03
    for: 10m
    labels:
      severity: critical
  - alert: K8S的APISERVER存在返回错误
    annotations:
      message: '{{ $labels.cluster }} 集群 API server 请求中有 {{ $value | humanizePercentage }} 的返回错误！'
    expr: |
      sum(rate(apiserver_request_total{job="apiserver",code=~"5.."}[5m]))
        /
      sum(rate(apiserver_request_total{job="apiserver"}[5m])) > 0.01
    for: 10m
    labels:
      severity: warning
  - alert: K8S的APISERVER资源存在返回错误过高
    annotations:
      message: '{{ $labels.cluster }} 集群 API server 的 {{ $labels.verb }} {{ $labels.resource }} {{ $labels.subresource }} 的请求中有 {{ $value | humanizePercentage }} 的返回错误！'
    expr: |
      sum(rate(apiserver_request_total{job="apiserver",code=~"5.."}[5m])) by (resource,subresource,verb,cluster)
        /
      sum(rate(apiserver_request_total{job="apiserver"}[5m])) by (resource,subresource,verb,cluster) > 0.10
    for: 10m
    labels:
      severity: critical
  - alert: K8S的APISERVER资源存在返回错误
    annotations:
      message: '{{ $labels.cluster }} 集群 API server 的 {{ $labels.verb }} {{ $labels.resource }} {{ $labels.subresource }} 的请求中有 {{ $value | humanizePercentage }} 的返回错误！'
    expr: |
      sum(rate(apiserver_request_total{job="apiserver",code=~"5.."}[5m])) by (resource,subresource,verb,cluster)
        /
      sum(rate(apiserver_request_total{job="apiserver"}[5m])) by (resource,subresource,verb,cluster) > 0.05
    for: 10m
    labels:
      severity: warning
  - alert: K8S客户端证书即将过期
    annotations:
      message: '{{ $labels.cluster }} 集群一个 K8S 的客户端证书将在 7 天内过期！'
    expr: |
      apiserver_client_certificate_expiration_seconds_count{job="apiserver"} > 0 and histogram_quantile(0.01, sum by (job, le) (rate(apiserver_client_certificate_expiration_seconds_bucket{job="apiserver"}[5m]))) < 604800
    labels:
      severity: warning
  - alert: K8S客户端证书24小时内过期
    annotations:
      message: '{{ $labels.cluster }} 集群一个 K8S 的客户端证书将在 24 小时内过期！'
    expr: |
      apiserver_client_certificate_expiration_seconds_count{job="apiserver"} > 0 and histogram_quantile(0.01, sum by (job, le) (rate(apiserver_client_certificate_expiration_seconds_bucket{job="apiserver"}[5m]))) < 86400
    labels:
      severity: critical
  - alert: APISERVER掉线
    annotations:
      message: '{{ $labels.cluster }} 集群 Prometheus Targets 无法发现 APISERVER！'
    expr: |
      absent(up{job="apiserver"} == 1)
    for: 15m
    labels:
      severity: critical
```

## 3.  apps



```yaml
groups:
- name: kubernetes-apps
  rules:
  - alert: K8S容器组短时间内多次重启
    annotations:
      message: '{{ $labels.cluster }} 集群容器组 {{ $labels.namespace }}/{{ $labels.pod }} ({{ $labels.container }}) 在10分钟内重启了 {{ printf "%.2f" $value }} 次！'
    expr: |
      rate(kube_pod_container_status_restarts_total{job="kube-state-metrics"}[15m]) * 60 * 10 > 1
    for: 10m
    labels:
      severity: critical
#  - alert: K8S容器组Terminated
#    annotations:
#      message: '{{ $labels.cluster }} 集群容器组 {{ $labels.namespace }}/{{ $labels.pod }} Terminated 原因是 {{ $labels.reason }}！'
#    expr: |
#      kube_pod_container_status_terminated_reason{reason!="Completed"} > 0
#    for: 15m
#    labels:
#      severity: warning
#  - alert: K8S容器组Completed
#    annotations:
#      message: '{{ $labels.cluster }} 集群容器组 {{ $labels.namespace }}/{{ $labels.pod }} Terminated 原因是 {{ $labels.reason }}！'
#    expr: |
#      kube_pod_container_status_terminated_reason{reason="Completed"} > 0
#    for: 15m
#    labels:
#      severity: none
  - alert: K8S容器组Waiting
    annotations:
      message: '{{ $labels.cluster }} 集群容器组 {{ $labels.namespace }}/{{ $labels.pod }} Waiting 原因是 {{ $labels.reason }}！'
    expr: |
      kube_pod_container_status_waiting_reason{reason!="ContainerCreating"} > 0
    for: 3m
    labels:
      severity: critical
  - alert: K8S容器组调度失败
    annotations:
      message: '{{ $labels.cluster }} 集群容器组 {{ $labels.namespace }}/{{ $labels.pod }} 无符合预期工作节点，无法被调度！'
    expr: |
      sum by (cluster,pod) (kube_pod_status_unschedulable) > 0
    for: 5m
    labels:
      severity: critical
  - alert: K8S容器组NotReady
    annotations:
      message: '{{ $labels.cluster }} 集群 {{ $labels.namespace }}/{{ $labels.pod }} 已处于 non-ready 状态超过15分钟！'
    expr: |
      sum by (namespace, pod, cluster) (max by(namespace, pod, cluster) (kube_pod_status_phase{job="kube-state-metrics", phase=~"Pending|Unknown"}) * on(namespace, pod, cluster) group_left(owner_kind) max by(namespace, pod, owner_kind, cluster) (kube_pod_owner{owner_kind!="Job"})) > 0
    for: 15m
    labels:
      severity: critical
  - alert: K8S部署状态异常
    annotations:
      message: '{{ $labels.cluster }} 集群部署的 {{ $labels.namespace }}/{{ $labels.deployment }} 状态异常，部分实例不可用已达15分钟！'
    expr: |
      kube_deployment_status_replicas_unavailable{cluster="prod"} != 0
    for: 15m
    labels:
      severity: warning
  - alert: K8S部署版本号不匹配
    annotations:
      message: '{{ $labels.cluster }} 集群部署的 {{ $labels.namespace }}/{{ $labels.deployment }} 部署版本号不匹配，这表明部署的部署过程失败，并且没有回滚达15分钟！'
    expr: |
      kube_deployment_status_observed_generation{job="kube-state-metrics"}
        !=
      kube_deployment_metadata_generation{job="kube-state-metrics"}
    for: 15m
    labels:
      severity: critical
  - alert: K8S部署实际副本数与预期数不匹配
    annotations:
      message: '{{ $labels.cluster }} 集群部署 {{ $labels.namespace }}/{{ $labels.deployment }} 部署的实际副本数与预期数不匹配超过15分钟！'
    expr: |
      kube_deployment_spec_replicas{job="kube-state-metrics"}
        !=
      kube_deployment_status_replicas_available{job="kube-state-metrics"}
    for: 15m
    labels:
      severity: critical
  - alert: K8S有状态部署实际副本数与预期数不匹配
    annotations:
      message: '{{ $labels.cluster }} 集群有状态部署 {{ $labels.namespace }}/{{ $labels.deployment }} 有状态部署的实际副本数与预期数不匹配超过15分钟！'
    expr: |
      kube_statefulset_status_replicas_ready{job="kube-state-metrics"}
        !=
      kube_statefulset_status_replicas{job="kube-state-metrics"}
    for: 15m
    labels:
      severity: critical
  - alert: K8S有状态部署版本号不匹配
    annotations:
      message: '{{ $labels.cluster }} 集群有状态部署的 {{ $labels.namespace }}/{{ $labels.deployment }} 有状态部署版本号不匹配，这表明有状态部署状态失败，并且没有回滚！'
    expr: |
      kube_statefulset_status_observed_generation{job="kube-state-metrics"}
        !=
      kube_statefulset_metadata_generation{job="kube-state-metrics"}
    for: 15m
    labels:
      severity: critical
  - alert: K8S有状态部署更新未展开
    annotations:
      message: '{{ $labels.cluster }} 集群有状态部署 {{ $labels.namespace }}/{{ $labels.statefulset }} 的更新未展开，发现当前本非更新版本！'
    expr: |
      max without (revision) (
        kube_statefulset_status_current_revision{job="kube-state-metrics"}
          unless
        kube_statefulset_status_update_revision{job="kube-state-metrics"}
      )
        *
      (
        kube_statefulset_replicas{job="kube-state-metrics"}
          !=
        kube_statefulset_status_replicas_updated{job="kube-state-metrics"}
      )
    for: 15m
    labels:
      severity: critical
  - alert: K8S守护进程集展开失败
    annotations:
      message: '{{ $labels.cluster }} 集群守护进程集 {{ $labels.namespace }}/{{ $labels.daemonset }} 只有预期容器组数的 {{ $value | humanizePercentage }} 的容器被调度并就绪！'
    expr: |
      kube_daemonset_status_number_ready{job="kube-state-metrics"}
        /
      kube_daemonset_status_desired_number_scheduled{job="kube-state-metrics"} < 1.00
    for: 15m
    labels:
      severity: critical
#  - alert: K8S容器等待中
#    annotations:
#      message: '{{ $labels.cluster }} 集群容器组 {{ $labels.namespace }}/{{ $labels.pod }} 中的 {{ $labels.container}} 容器已经再等待状态超过1小时！'
#    expr: |
#      sum by (cluster, namespace, pod, container) (kube_pod_container_status_waiting_reason{job="kube-state-metrics"}) > 0
#    for: 1h
#    labels:
#      severity: warning
  - alert: K8S守护进程集未被调度
    annotations:
      message: '{{ $labels.cluster }} 集群守护进程集 {{ $labels.namespace }}/{{ $labels.daemonset }} 的 {{ $value }} 个容器组没有被调度！'
    expr: |
      kube_daemonset_status_desired_number_scheduled{job="kube-state-metrics"}
        -
      kube_daemonset_status_current_number_scheduled{job="kube-state-metrics"} > 0
    for: 10m
    labels:
      severity: warning
  - alert: K8S守护进程集调度错误
    annotations:
      message: '{{ $labels.cluster }} 集群守护进程集 {{ $labels.namespace }}/{{ $labels.daemonset }} 的 {{ $value }} 个非预期的容器组正在运行！'
    expr: |
      kube_daemonset_status_number_misscheduled{job="kube-state-metrics"} > 0
    for: 10m
    labels:
      severity: warning
  - alert: K8S定时任务运行中
    annotations:
      message: '{{ $labels.cluster }} 集群定时任务 {{ $labels.namespace }}/{{ $labels.cronjob }} 已经使用1小时时间来完成任务！'
    expr: |
      time() - kube_cronjob_next_schedule_time{job="kube-state-metrics"} > 3600
    for: 1h
    labels:
      severity: warning
  - alert: K8S任务完成
    annotations:
      message: '{{ $labels.cluster }} 集群任务 {{ $labels.namespace }}/{{ $labels.cronjob }} 已经使用1小时时间来完成任务！'
    expr: |
      kube_job_spec_completions{job="kube-state-metrics"} - kube_job_status_succeeded{job="kube-state-metrics"}  > 0
    for: 1h
    labels:
      severity: warning
  - alert: K8S任务失败
    annotations:
      message: '{{ $labels.cluster }} 集群任务 {{ $labels.namespace }}/{{ $labels.cronjob }} 已经失败！'
    expr: |
      kube_job_failed{job="kube-state-metrics"}  > 0
    for: 15m
    labels:
      severity: warning
  - alert: K8S的HPA副本数不匹配
    annotations:
      message: '{{ $labels.cluster }} 集群HPA {{ $labels.namespace }}/{{ $labels.hpa }} 与预期副本数不匹配已经超过15分钟！'
    expr: |
      (kube_hpa_status_desired_replicas{job="kube-state-metrics"}
        !=
      kube_hpa_status_current_replicas{job="kube-state-metrics"})
        and
      changes(kube_hpa_status_current_replicas[15m]) == 0
    for: 15m
    labels:
      severity: warning
  - alert: 侦测到K8S的HPA缩容
    annotations:
      message: '{{ $labels.cluster }} 集群 HPA {{ $labels.namespace }}/{{ $labels.hpa }} 已触发缩容，可用副本数达到预期，当前预期 {{ printf "%.0f" $value }} ！'
    expr: |
      (kube_hpa_status_desired_replicas{job="kube-state-metrics"}
        ==
      kube_hpa_status_current_replicas{job="kube-state-metrics"})
        and
      delta(kube_hpa_status_current_replicas[5m]) < 0
    for: 1m
    labels:
      severity: none
  - alert: 侦测到K8S的HPA扩容
    annotations:
      message: '{{ $labels.cluster }} 集群 HPA {{ $labels.namespace }}/{{ $labels.hpa }} 已触发扩容，可用副本数达到预期，当前预期 {{ printf "%.0f" $value }} ！！'
    expr: |
      (kube_hpa_status_desired_replicas{job="kube-state-metrics"}
        ==
      kube_hpa_status_current_replicas{job="kube-state-metrics"})
        and
      delta(kube_hpa_status_current_replicas[5m]) > 0
    for: 1m
    labels:
      severity: none
  - alert: K8S工作负载的HPA保持满载
    annotations:
      message: '{{ $labels.cluster }} 集群 HPA {{ $labels.namespace }}/{{ $labels.hpa }} 以限制最大副本数满载运行超过了15分钟！'
    expr: |
      kube_hpa_status_current_replicas{job="kube-state-metrics"}
        ==
      kube_hpa_spec_max_replicas{job="kube-state-metrics"}
    for: 15m
    labels:
      severity: none
  - alert: K8S部署服务版本变更通告
    annotations:
      message: '侦测到 {{ $labels.cluster }} 集群服务部署 {{ $labels.namespace }}/{{ $labels.deployment }} 部署 metadata 版本已更替，实列数以达到预设值。'
    expr: |
      (kube_deployment_status_observed_generation{job="kube-state-metrics"}
        ==
      kube_deployment_metadata_generation{job="kube-state-metrics"})
        and
      (kube_deployment_spec_replicas{job="kube-state-metrics"}
        ==
      kube_deployment_status_replicas_available{job="kube-state-metrics"})
        and
      changes(kube_deployment_status_observed_generation{job="kube-state-metrics"}[5m]) > 0
    for: 1m
    labels:
      severity: none
  - alert: K8S部署服务版本变更异常
    annotations:
      message: '侦测到 {{ $labels.cluster }} 集群服务部署 {{ $labels.namespace }}/{{ $labels.deployment }} 部署 metadata 版本已更替，实列在线数不匹配部署预设值，当前运行版本非新版本，或 HPA 已触发，或服务运行故障！'
    expr: |
     ((kube_deployment_status_observed_generation{job="kube-state-metrics"}
        !=
      kube_deployment_metadata_generation{job="kube-state-metrics"})
        or
      (kube_deployment_spec_replicas{job="kube-state-metrics"}
        !=
      kube_deployment_status_replicas_available{job="kube-state-metrics"}))
        or
      ((kube_hpa_status_desired_replicas{job="kube-state-metrics"}
        !=
      kube_hpa_status_current_replicas{job="kube-state-metrics"})
        and
      changes(kube_hpa_status_current_replicas[15m]) != 0)
        and
      changes(kube_deployment_status_observed_generation{job="kube-state-metrics"}[5m]) > 0
    for: 1m
    labels:
      severity: critical
```



## 4.  controller-manager



```yaml
groups:
- name: kubernetes-system-controller-manager
  rules:
  - alert: KubeControllerManager掉线
    annotations:
      message: KubeControllerManager 从 Prometheus Targets 的发现中消失！
    expr: |
      absent(up{job="kube-controller-manager"} == 1)
    for: 15m
    labels:
      severity: critical
```

## 5. general



```yaml
groups:
- name: general.rules
  rules:
  - alert: Target掉线
    annotations:
      message: '{{ $labels.cluster }} 集群命名空间 {{ $labels.namespace }} 中 {{ $labels.job }} 的 {{ printf "%.4g" $value }}% 个targets掉线！'
    expr: 100 * (count(up == 0) BY (cluster, job, namespace, service) / count(up) BY (cluster, job,
      namespace, service)) > 10
    for: 10m
    labels:
      severity: warning
  - alert: Watchdog
    annotations:
      message: |
        此警报旨在确认整个警报管道功能性的。这个警报始终处于触发状态，因此它应始终在Alertmanager中触发，并始终针对各类接收器发送。
    expr: vector(1)
    labels:
      severity: none
```



## 6. kubelet



```yaml
groups:
- name: kubernetes-system-kubelet
  rules:
  - alert: K8S节点未就绪
    annotations:
      message: '{{ $labels.cluster }} 集群K8S节点 {{ $labels.node }} 处于未就绪状态已超过15分钟！'
    expr: |
      kube_node_status_condition{job="kube-state-metrics",condition="Ready",status="true"} == 0
    for: 15m
    labels:
      severity: warning
  - alert: K8S节点不可达
    annotations:
      message: '{{ $labels.cluster }} 集群K8S节点 {{ $labels.node }} 不可达，一部分工作负载已重新调度！'
    expr: |
      kube_node_spec_taint{job="kube-state-metrics",key="node.kubernetes.io/unreachable",effect="NoSchedule"} == 1
    labels:
      severity: warning
  - alert: Kubelet节点存在过多容器组
    annotations:
      message: '{{ $labels.cluster }} 集群 Kubelet {{ $labels.node }} 节点已经运行了其总量的 {{ $value | humanizePercentage }} 的容器组再这个节点上！'
    expr: |
      max(max(kubelet_running_pod_count{job="kubelet", metrics_path="/metrics"}) by(instance,cluster) * on(instance,cluster) group_left(node) kubelet_node_name{job="kubelet", metrics_path="/metrics"}) by(node,cluster) / max(kube_node_status_capacity_pods{job="kube-state-metrics"}) by(node,cluster) > 0.95
    for: 15m
    labels:
      severity: warning
  - alert: Kubelet掉线
    annotations:
      message: '{{ $labels.cluster }} 集群 Prometheus Targets 无法发现 Kubelet {{ $labels.node }}！'
    expr: |
      absent(up{job="kubelet", metrics_path="/metrics"} == 1)
    for: 15m
    labels:
      severity: critical
```

## 7. network

```yaml
groups:
- name: node-network
  rules:
  - alert: Node网络网卡抖动
    annotations:
      message: '{{ $labels.cluster }} 集群侦测到 node-exporter {{ $labels.namespace }}/{{ $labels.pod }} 节点上的网卡 {{ $labels.device }} 状态经常改变！'
    expr: |
      changes(node_network_up{job="node-exporter",device!~"veth.+"}[2m]) > 2
    for: 2m
    labels:
      severity: warning
#  - alert: 节点侦测到TCP已分配的套接字数量
#    expr: sum(avg_over_time(node_sockstat_TCP_alloc[5m])) by (instance,cluster)  > 5000
#    for: 1m
#    labels:
#      severity: critical
#    annotations:
#      message: '{{ $labels.cluster }} 集群 {{ $labels.instance }} 节点侦测到 TCP 已分配的套接字数量达到 {{ printf "%.0f" $value }}!'
#  - alert: 节点侦测到UDP使用中的套接字数量
#    expr: sum(avg_over_time(node_sockstat_UDP_inuse[5m])) by (instance,cluster)  > 5000
#    for: 1m
#    labels:
#      severity: critical
#    annotations:
#      message: '{{ $labels.cluster }} 集群 {{ $labels.instance }} 节点侦测到UDP使用中的套接字数量达到 {{ printf "%.0f" $value }}!'
  - alert: 节点下行网络错误
    annotations:
      message: '{{ $labels.cluster }} 集群 {{ $labels.instance }} 节点的网络设备 {{ $labels.device }} 再过去2分钟内侦测到 {{ printf "%.0f" $value }} 的下载错误！'
    expr: |
      increase(node_network_receive_errs_total[2m]) > 10
    for: 5m
    labels:
      severity: warning
  - alert: 节点上行网络错误
    annotations:
      message: '{{ $labels.cluster }} 集群 {{ $labels.instance }} 节点的网络设备 {{ $labels.device }} 再过去2分钟内侦测到 {{ printf "%.0f" $value }} 的上传错误！'
    expr: |
      increase(node_network_transmit_errs_total[2m]) > 10
    for: 5m
    labels:
      severity: warning
  - alert: 节点下行带宽过高
    annotations:
      message: '{{ $labels.cluster }} 集群 {{ $labels.instance }} 节点的网络设备 {{ $labels.device }} 下载带宽超过 > 100MB/s'
    expr: |
      sum by (icluster,instance) (irate(node_network_receive_bytes_total[2m])) / 1024 / 1024 > 100
    for: 5m
    labels:
      severity: warning
  - alert: 节点上行带宽过高
    annotations:
      message: '{{ $labels.cluster }} 集群 {{ $labels.instance }} 节点的网络设备 {{ $labels.device }} 上传带宽超过 > 100MB/s'
    expr: |
      sum by (cluster,instance) (irate(node_network_transmit_bytes_total[2m])) / 1024 / 1024 > 100
    for: 5m
    labels:
      severity: warning
  - alert: 节点下行丢包率过高
    annotations:
      message: '{{ $labels.cluster }} 集群 {{ $labels.instance }} 节点3分钟内下行丢包率超过达到 {{ printf "%.0f" $value }}%！'
    expr: |
      sum by (instance,cluster) (irate(node_network_receive_drop_total[3m])) / sum by (instance,cluster) (irate(node_network_receive_packets_total[3m])) * 100 > 80
    for: 1m
    labels:
      severity: cirtical
  - alert: 节点上行丢包率过高
    annotations:
      message: '{{ $labels.cluster }} 集群 {{ $labels.instance }} 节点3分钟内上行丢包率超过达到 {{ printf "%.0f" $value }}%！'
    expr: |
      sum by (instance,cluster) (irate(node_network_transmit_drop_total[3m])) / sum by (instance,cluster) (irate(node_network_transmit_packets_total[3m])) * 100 > 80
    for: 1m
    labels:
      severity: cirtical
```

## 8. prometheus-operator

```yaml
groups:
- name: prometheus-operator
  rules:
  - alert: PrometheusOperatorReconcileErrors
    annotations:
      message: '{{ $labels.cluster }} 集群 {{ $labels.namespace }} 命名空间中协调 {{ $labels.controller }} 时发生错误！'
    expr: |
      rate(prometheus_operator_reconcile_errors_total{job="prometheus-operator",namespace="monitoring"}[5m]) > 0.1
    for: 10m
    labels:
      severity: warning
  - alert: PrometheusOperator节点lookup错误
    annotations:
      message: '{{ $labels.cluster }} 集群协调 Prometheus 时 {{ $labels.namespace }} 命名空间发生错误！'
    expr: |
      rate(prometheus_operator_node_address_lookup_errors_total{job="prometheus-operator",namespace="monitoring"}[5m]) > 0.1
    for: 10m
    labels:
      severity: warning
```

## 9. prometheus



```yaml
groups:
- name: prometheus
  rules:
  - alert: Prometheus错误的配置
    annotations:
      message: '{{ $labels.cluster }} 集群 Prometheus {{$labels.namespace}}/{{$labels.pod}} 再重载配置时失败！'
    expr: |
      # Without max_over_time, failed scrapes could create false negatives, see
      # https://www.robustperception.io/alerting-on-gauges-in-prometheus-2-0 for details.
      max_over_time(prometheus_config_last_reload_successful{job="prometheus-k8s",namespace="monitoring"}[5m]) == 0
    for: 10m
    labels:
      severity: critical
  - alert: Prometheus通知队列已满
    annotations:
      message: Prometheus {{$labels.namespace}}/{{$labels.pod}} 的报警通知队列已满！
        30m.
    expr: |
      # Without min_over_time, failed scrapes could create false negatives, see
      # https://www.robustperception.io/alerting-on-gauges-in-prometheus-2-0 for details.
      (
        predict_linear(prometheus_notifications_queue_length{job="prometheus-k8s",namespace="monitoring"}[5m], 60 * 30)
      >
        min_over_time(prometheus_notifications_queue_capacity{job="prometheus-k8s",namespace="monitoring"}[5m])
      )
    for: 15m
    labels:
      severity: warning
  - alert: Prometheus在推送警报时发生错误
    annotations:
      message: '{{ $labels.cluster }} 集群 {{$labels.namespace}}/{{$labels.pod}} 在推送警报至某些 Alertmanager {{$labels.alertmanager}} 时出现了 {{ printf "%.1f" $value }}% 的错误！'
    expr: |
      (
        rate(prometheus_notifications_errors_total{job="prometheus-k8s",namespace="monitoring"}[5m])
      /
        rate(prometheus_notifications_sent_total{job="prometheus-k8s",namespace="monitoring"}[5m])
      )
      * 100
      > 1
    for: 15m
    labels:
      severity: warning
  - alert: Prometheus在推送警报时全部错误
    annotations:
      message: '{{ $labels.cluster }} 集群 {{$labels.namespace}}/{{$labels.pod}} 在推送警报至全部 Alertmanager {{$labels.alertmanager}} 时出现了 {{ printf "%.1f" $value }}% 的错误！'
    expr: |
      min without(alertmanager) (
        rate(prometheus_notifications_errors_total{job="prometheus-k8s",namespace="monitoring"}[5m])
      /
        rate(prometheus_notifications_sent_total{job="prometheus-k8s",namespace="monitoring"}[5m])
      )
      * 100
      > 3
    for: 15m
    labels:
      severity: critical
  - alert: Prometheus未连接Alertmanagers
    annotations:
      message: '{{ $labels.cluster }} 集群 Prometheus {{$labels.namespace}}/{{$labels.pod}} 没有连接到任何 Alertmanagers！'
    expr: |
      max_over_time(prometheus_notifications_alertmanagers_discovered{job="prometheus"}[5m]) < 1
    for: 10m
    labels:
      severity: warning
  - alert: PrometheusTSDB重载失败
    annotations:
      message: '{{ $labels.cluster }} 集群在过去的3小时内 Prometheus {{$labels.namespace}}/{{$labels.pod}} 侦测到 {{$value | humanize}} 个重载错误！'
    expr: |
      increase(prometheus_tsdb_reloads_failures_total{job="prometheus-k8s",namespace="monitoring"}[3h]) > 0
    for: 4h
    labels:
      severity: warning
  - alert: PrometheusTSDB压缩失败
    annotations:
      message: '{{ $labels.cluster }} 集群在过去的3小时内 Prometheus {{$labels.namespace}}/{{$labels.pod}} has detected {{$value | humanize}} 个压缩错误！'
    expr: |
      increase(prometheus_tsdb_compactions_failed_total{job="prometheus-k8s",namespace="monitoring"}[3h]) > 0
    for: 4h
    labels:
      severity: warning
  - alert: Prometheus没有采集到数据样本
    annotations:
      message: '{{ $labels.cluster }} 集群 Prometheus {{$labels.namespace}}/{{$labels.pod}} 没有采集到数据样本！'
    expr: |
      rate(prometheus_tsdb_head_samples_appended_total{job="prometheus-k8s",namespace="monitoring"}[5m]) <= 0
    for: 10m
    labels:
      severity: warning
  - alert: Prometheus重复的时间戳
    annotations:
      message: '{{ $labels.cluster }} 集群 Prometheus {{$labels.namespace}}/{{$labels.pod}} 正在丢弃 {{ printf "%.4g" $value  }} 拥有相同时间戳不同数据的数据样本！'
    expr: |
      rate(prometheus_target_scrapes_sample_duplicate_timestamp_total{job="prometheus-k8s",namespace="monitoring"}[5m]) > 0
    for: 10m
    labels:
      severity: warning
  - alert: Prometheus时间戳超过限制
    annotations:
      message: '{{ $labels.cluster }} 集群 Prometheus {{$labels.namespace}}/{{$labels.pod}} 正在丢弃 {{ printf "%.4g" $value  }} 超过时间戳限制的数据样本！'
    expr: |
      rate(prometheus_target_scrapes_sample_out_of_order_total{job="prometheus-k8s",namespace="monitoring"}[5m]) > 0
    for: 10m
    labels:
      severity: warning
  - alert: Prometheus远程存储失败
    annotations:
      message: '{{ $labels.cluster }} 集群 Prometheus {{$labels.namespace}}/{{$labels.pod}} 在推送至数据都队列 {{$labels.queue}} 数据时有 {{ printf "%.1f" $value }}% 的错误！'
    expr: |
      (
        rate(prometheus_remote_storage_failed_samples_total{job="prometheus-k8s",namespace="monitoring"}[5m])
      /
        (
          rate(prometheus_remote_storage_failed_samples_total{job="prometheus-k8s",namespace="monitoring"}[5m])
        +
          rate(prometheus_remote_storage_succeeded_samples_total{job="prometheus-k8s",namespace="monitoring"}[5m])
        )
      )
      * 100
      > 1
    for: 15m
    labels:
      severity: critical
  - alert: Prometheus远程数据写落后
    annotations:
      message: '{{ $labels.cluster }} 集群 Prometheus {{$labels.namespace}}/{{$labels.pod}} 远程写落后于队列 {{$labels.queue}} {{ printf "%.1f" $value }} 秒！'
    expr: |
      # Without max_over_time, failed scrapes could create false negatives, see
      # https://www.robustperception.io/alerting-on-gauges-in-prometheus-2-0 for details.
      (
        max_over_time(prometheus_remote_storage_highest_timestamp_in_seconds{job="prometheus-k8s",namespace="monitoring"}[5m])
      - on(job, instance) group_right
        max_over_time(prometheus_remote_storage_queue_highest_sent_timestamp_seconds{job="prometheus-k8s",namespace="monitoring"}[5m])
      )
      > 120
    for: 15m
    labels:
      severity: critical
  - alert: Prometheus远程写预期切片
    annotations:
      message: '{{ $labels.cluster }} 集群 Prometheus {{$labels.namespace}}/{{$labels.pod}} 远程写的预期切片数估计需要 {{ $value }} shards, 大于最大值 {{ printf `prometheus_remote_storage_shards_max{instance="%s",job="prometheus-k8s",namespace="monitoring"}` $labels.instance | query | first | value }}！'
    expr: |
      # Without max_over_time, failed scrapes could create false negatives, see
      # https://www.robustperception.io/alerting-on-gauges-in-prometheus-2-0 for details.
      (
        max_over_time(prometheus_remote_storage_shards_desired{job="prometheus-k8s",namespace="monitoring"}[5m])
      >
        max_over_time(prometheus_remote_storage_shards_max{job="prometheus-k8s",namespace="monitoring"}[5m])
      )
    for: 15m
    labels:
      severity: warning
  - alert: Prometheus规则错误
    annotations:
      message: '{{ $labels.cluster }} 集群在5分钟内 Prometheus {{$labels.namespace}}/{{$labels.pod}} 评估 {{ printf "%.0f" $value }} 条的规则失败！'
    expr: |
      increase(prometheus_rule_evaluation_failures_total{job="prometheus-k8s",namespace="monitoring"}[5m]) > 0
    for: 15m
    labels:
      severity: critical
  - alert: Prometheus缺少规则评估
    annotations:
      message: '{{ $labels.cluster }} 集群在过去5分钟内 Prometheus {{$labels.namespace}}/{{$labels.pod}} 错过了 {{ printf "%.0f" $value }} 规则组评估！'
    expr: |
      increase(prometheus_rule_group_iterations_missed_total{job="prometheus-k8s",namespace="monitoring"}[5m]) > 0
    for: 15m
    labels:
      severity: warning
```

## 10. resource

```yaml
groups:     
- name: kubernetes-resources
  rules:
  - alert: K8S的CPU的Requests过载
    annotations:
      message: '{{ $labels.cluster }} 群集对容器组的 CPU 资源 Requests 过载，并且无容忍策略，集群需要扩容！'
    expr: |
      sum(namespace:kube_pod_container_resource_requests_cpu_cores:sum)
        /
      sum(kube_node_status_allocatable_cpu_cores)
        >
      (count(kube_node_status_allocatable_cpu_cores)-1) / count(kube_node_status_allocatable_cpu_cores)
    for: 5m
    labels:
      severity: warning
  - alert: K8S的内存Requests过载
    annotations:
      message: '{{ $labels.cluster }} 群集对容器组的内存资源Requests过载，并且无容忍策略，集群需要扩容！'
    expr: |
      sum(namespace:kube_pod_container_resource_requests_memory_bytes:sum)
        /
      sum(kube_node_status_allocatable_memory_bytes)
        >
      (count(kube_node_status_allocatable_memory_bytes)-1)
        /
      count(kube_node_status_allocatable_memory_bytes)
    for: 5m
    labels:
      severity: warning
  - alert: K8S工作节点的CPURequests过载
    annotations:
      message: '{{ $labels.cluster }} 集群容器组对节点 {{ $labels.node }} 的 CPU 资源 Requests 以达到 {{ printf "%.0f" $value }}%！'
    expr: |
      sum by (node,cluster) (kube_pod_container_resource_requests_cpu_cores) / sum by (node,cluster) (node:node_num_cpu:sum) * 100 > 95
    for: 5m
    labels:
      severity: warning
  - alert: K8S工作节点的平均CPURequests过载
    annotations:
      message: '{{ $labels.cluster }} 集群容器组对节点 {{ $labels.node }} 的 CPU 资源平均 Requests 以达到 {{ printf "%.0f" $value }}%，可能导致无法调度，{{ $labels.cluster }} 集群可能需要扩容！'
    expr: |
      avg by (cluster) (sum by (node,cluster) (kube_pod_container_resource_requests_cpu_cores) / sum by (node,cluster) (node:node_num_cpu:sum)) * 100 > 90
    for: 5m
    labels:
      severity: warning
  - alert: K8S工作节点内存Requests过载
    annotations:
      message: '{{ $labels.cluster }} 集群容器组对节点 {{ $labels.node }} 的内存资源 Requests 以达到 {{ printf "%.0f" $value }}%！'
    expr: |
      sum by (node,cluster) (kube_pod_container_resource_requests_memory_bytes) / sum by (node,cluster) (kube_node_status_allocatable_memory_bytes) * 100 > 95
    labels:
      severity: warning
  - alert: K8S工作节点平均内存Requests过载
    annotations:
      message: '{{ $labels.cluster }} 集群容器组对节点 {{ $labels.node }} 的内存资源平均 Requests 以达到 {{ printf "%.0f" $value }}%，可能导致无法调度，{{ $labels.cluster }} 集群可能需要扩容！'
    expr: |
      avg by (cluster) (sum by (node,cluster) (kube_pod_container_resource_requests_memory_bytes) / sum by (node,cluster) (kube_node_status_allocatable_memory_bytes)) * 100 > 85
    labels:
      severity: warning
  - alert: 'K8S的命名空间CPU过载'
    annotations:
      message: '{{ $labels.cluster }} 集群命名空间的CPU过载！'
    expr: |
      sum(kube_resourcequota{job="kube-state-metrics", type="hard", resource="cpu"})
        /
      sum(kube_node_status_allocatable_cpu_cores)
        > 1.5
    for: 5m
    labels:
      severity: warning
  - alert: K8S的命名空间内存过载
    annotations:
      message: '{{ $labels.cluster }} 集群命名空间的内存过载！'
    expr: |
      sum(kube_resourcequota{job="kube-state-metrics", type="hard", resource="memory"})
        /
      sum(kube_node_status_allocatable_memory_bytes{job="node-exporter"})
        > 1.5
    for: 5m
    labels:
      severity: warning
  - alert: K8S超过配额
    annotations:
      message: '{{ $labels.cluster }} 集群命名空间 {{ $labels.namespace }} 已使用了其配额的 {{ $labels.resource }} {{ $value | humanizePercentage }}！'
    expr: |
      kube_resourcequota{job="kube-state-metrics", type="used"}
        / ignoring(instance, job, type)
      (kube_resourcequota{job="kube-state-metrics", type="hard"} > 0)
        > 0.90
    for: 15m
    labels:
      severity: warning
  - alert: 有受限的CPU(CPU节流)
    annotations:
      message: '{{ $labels.cluster }} 集群命名空间 {{ $labels.namespace }} 的容器组 {{ $labels.pod }} 中的容器 {{ $labels.container }} 存在 {{ $value | humanizePercentage }}  受限 CPU（CPU节流）！'
    expr: |
      sum(increase(container_cpu_cfs_throttled_periods_total{container!="", }[5m])) by (container, pod, namespace,cluster)
        /
      sum(increase(container_cpu_cfs_periods_total{}[5m])) by (container, pod, namespace,cluster)
        > ( 100 / 100 )
    for: 15m
    labels:
      severity: warning
  - alert: K8S容器组CPULimits%使用率高
    annotations:
      message: '{{ $labels.cluster }} 集群命名空间 {{ $labels.namespace }} 的容器组 {{ $labels.pod }} 中的容器 {{ $labels.container }} CPU Limits %达到 {{ $value | humanizePercentage }} ！'
    expr: |
      sum(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate) by (container,pod,namespace,cluster) / sum(kube_pod_container_resource_limits_cpu_cores) by (container,pod,namespace,cluster) > 1
    for: 15m
    labels:
      severity: warning
  - alert: K8S容器组内存Limits%使用率高 
    annotations:
      message: '{{ $labels.cluster }} 集群命名空间 {{ $labels.namespace }} 的容器组 {{ $labels.pod }} 中的容器 {{ $labels.container }} 内存 Limits% 达到 {{ $value | humanizePercentage }} ！'
    expr: |
      sum(container_memory_working_set_bytes) by (container,pod,namespace,cluster) / sum(kube_pod_container_resource_limits_memory_bytes) by (container,pod,namespace,cluster) > 1
    for: 15m
    labels:
      severity: warning
#  - alert: K8S工作负载的HPA保持满载并且资源平均利用率高
#    annotations:
#      message: '{{ $labels.cluster }} 集群 HPA {{ $labels.namespace }}/{{ $labels.hpa }} 以限制副本数满载运行超过了15分钟，并且资源平均利用率达 {{ $value }}% ,需要扩容！'
#    expr: |
#      kube_hpa_status_current_metrics_average_utilization > 95
#        and
#      kube_hpa_status_current_replicas{job="kube-state-metrics"}
#        ==
#      kube_hpa_spec_max_replicas{job="kube-state-metrics"}
#    for: 15m
#    labels:
#      severity: critical
  - alert: K8S工作负载CPULimits%使用率高
    annotations:
      message: '{{ $labels.cluster }} 集群命名空间 {{ $labels.namespace }} 的 {{ $labels.workload_type }} 工作负载 {{ $labels.workload }} CPU Limits% 达到 {{ $value | humanizePercentage }} 可能触发 HPA！'
    expr: |
      sum(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate * on(namespace,pod,cluster) group_left(workload, workload_type) mixin_pod_workload) by (workload,workload_type,cluster,namespace) / sum(kube_pod_container_resource_limits_cpu_cores * on(namespace,pod,cluster) group_left(workload, workload_type) mixin_pod_workload) by (workload, workload_type,cluster,namespace) > 3.5
    for: 15m
    labels:
      severity: warning
  - alert: K8S工作负载CPURequests%使用率达HPA扩容阈值
    annotations:
      message: '{{ $labels.cluster }} 集群命名空间 {{ $labels.namespace }} 的 {{ $labels.workload_type }} 工作负载 {{ $labels.workload }} CPU Requests% 达到 {{ $value | humanizePercentage }} 达到 HPA 扩容条件！'
    expr: |
      sum(
        node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate{namespace=~"prod|super", container!=""}
          * on(namespace,pod,cluster) group_left(workload, workload_type) mixin_pod_workload) by (workload,workload_type,cluster,namespace)
        /
      sum(
         kube_pod_container_resource_requests_cpu_cores{namespace=~"prod|super"} * on(namespace,pod,cluster) group_left(workload, workload_type) mixin_pod_workload) by (workload, workload_type,cluster,namespace) > 4
        and
      count(
        sum(
          node_namespace_pod_container:container_cpu_usage_seconds_total:sum_rate{namespace=~"prod|super", container!=""}
            * on(namespace,pod,cluster) group_left(workload, workload_type) mixin_pod_workload) by (workload,workload_type,pod,cluster,namespace)
          /
        sum(
          kube_pod_container_resource_requests_cpu_cores{namespace=~"prod|super"}
            * on(namespace,pod,cluster) group_left(workload, workload_type) mixin_pod_workload) by (workload, workload_type,pod,cluster,namespace)
          > 4) by (workload, workload_type,cluster,namespace)
        ==
      count(
        sum(
          kube_pod_container_resource_requests_cpu_cores{namespace=~"prod|super"}
        * on(namespace,pod)
          group_left(workload, workload_type) mixin_pod_workload{namespace=~"prod|super"}) by (workload, workload_type,pod,cluster,namespace)) by (workload, workload_type,cluster,namespace)
    for: 30s
    labels:
      severity: none
  - alert: K8S工作负载内存Limits%使用率高
    annotations:
      message: '{{ $labels.cluster }} 集群命名空间 {{ $labels.namespace }} 的 {{ $labels.workload_type }} 工作负载 {{ $labels.workload }} 内存 Limits% 达到 {{ $value | humanizePercentage }} 可能触发 HPA！'
    expr: |
      sum(container_memory_working_set_bytes * on(namespace,pod,container,cluster) group_left(workload, workload_type) mixin_pod_workload) by (workload, workload_type,cluster,namespace) / sum(kube_pod_container_resource_limits_memory_bytes * on(namespace,pod,cluster) group_left(workload, workload_type) mixin_pod_workload) by (workload, workload_type,cluster,namespace) > 1
    for: 15m 
    labels:
      severity: warning
  - alert: K8S工作负载内存Requests%使用率达HPA扩容阈值
    annotations:
      message: '{{ $labels.cluster }} 集群命名空间 {{ $labels.namespace }} 的 {{ $labels.workload_type }} 工作负载 {{ $labels.workload }} 内存 Requests% 达到 {{ $value | humanizePercentage }} 达到 HPA 扩容条件！'
    expr: |
      (sum(
        container_memory_working_set_bytes{namespace=~"prod|super", container!=""}
          * on(namespace,pod,cluster) group_left(workload, workload_type) mixin_pod_workload) by (workload,workload_type,cluster,namespace)
        /
      sum(
         kube_pod_container_resource_requests_memory_bytes{namespace=~"prod|super"} * on(namespace,pod,cluster) group_left(workload, workload_type) mixin_pod_workload) by (workload, workload_type,cluster,namespace) > 1.1)
        and
      ((count(
        sum(
          container_memory_working_set_bytes{namespace=~"prod|super", container!=""}
            * on(namespace,pod,cluster) group_left(workload, workload_type) mixin_pod_workload) by (workload,workload_type,pod,cluster,namespace) 
          / 
        sum(
          kube_pod_container_resource_requests_memory_bytes{namespace=~"prod|super"}
            * on(namespace,pod,cluster) group_left(workload, workload_type) mixin_pod_workload) by (workload, workload_type,pod,cluster,namespace) 
          > 1.1) by (workload, workload_type,cluster,namespace))
        == 
      (count(
        sum(
          kube_pod_container_resource_requests_memory_bytes{namespace=~"prod|super"}
        * on(namespace,pod)
          group_left(workload, workload_type) mixin_pod_workload{namespace=~"prod|super"}) by (workload, workload_type,pod,cluster,namespace)) by (workload, workload_type,cluster,namespace)))
    for: 30s
    labels:
      severity: none
```



## 11. scheduler



```yaml
groups:
- name: kubernetes-system-scheduler
  rules:
  - alert: K8SScheduler掉线
    annotations:
      message: KubeScheduler 从 Prometheus Targets 的发现中消失！
    expr: |
      absent(up{job="kube-scheduler"} == 1)
    for: 15m
    labels:
      severity: critical
```

## 12. storage



```yaml
group:
- name: kubernetes-storage
  rules:
  - alert: K8S的PV使用量警报
    annotations:
      message: '{{ $labels.cluster }} 集群命名空间 {{ $labels.namespace }} 中被PVC {{ $labels.persistentvolumeclaim }} 声明的的PV只剩下 {{ $value | humanizePercentage }} 空闲！'
    expr: |
      kubelet_volume_stats_available_bytes{job="kubelet", metrics_path="/metrics"}
        /
      kubelet_volume_stats_capacity_bytes{job="kubelet", metrics_path="/metrics"}
        < 0.03
    for: 1m
    labels:
      severity: critical
  - alert: KubePersistentVolumeFullInFourDays
    annotations:
      message: '{{ $labels.cluster }} 集群通过抽样计算，命名空间 {{ $labels.namespace }} 中被PVC {{ $labels.persistentvolumeclaim }} 声明的的PV将在4天内用尽，当前剩余 {{ $value | humanizePercentage }}！'
    expr: |
      (
        kubelet_volume_stats_available_bytes{job="kubelet", metrics_path="/metrics"}
          /
        kubelet_volume_stats_capacity_bytes{job="kubelet", metrics_path="/metrics"}
      ) < 0.15
      and
      predict_linear(kubelet_volume_stats_available_bytes{job="kubelet", metrics_path="/metrics"}[6h], 4 * 24 * 3600) < 0
    for: 1h
    labels:
      severity: critical
  - alert: K8S的PV错误
    annotations:
      message: '{{ $labels.cluster }} 集群 PV {{ $labels.persistentvolume }} 的状态为 {{ $labels.phase }}！'
    expr: |
      kube_persistentvolume_status_phase{phase=~"Failed|Pending",job="kube-state-metrics"} > 0
    for: 5m
    labels:
      severity: critical
```

## 13. system



```yaml
groups: 
- name: kubernetes-system
  rules:
  - alert: 节点文件系统24小时内用完
    annotations:
      message: '{{ $labels.cluster }} 集群的 {{ $labels.instance }} 节点的文件系统的 {{ $labels.device }} 设备只剩下 {{ printf "%.2f" $value }}% 可使用空间，速率计算可能在24小时内填满！'
    expr: |
      (
        node_filesystem_avail_bytes{job="node-exporter",fstype!=""} / node_filesystem_size_bytes{job="node-exporter",fstype!=""} * 100 < 40
      and
        predict_linear(node_filesystem_avail_bytes{job="node-exporter",fstype!=""}[6h], 24*60*60) < 0
      and
        node_filesystem_readonly{job="node-exporter",fstype!=""} == 0
      )
    for: 1h
    labels:
      severity: warning
  - alert: 节点文件系统4小时内用完
    annotations:
      message: '{{ $labels.cluster }} 集群的 {{ $labels.instance }} 节点的文件系统的 {{ $labels.device }} 设备只剩下 {{ printf "%.2f" $value }}% 可使用空间，速率计算可能在4小时内填满！'
    expr: |
      (
        node_filesystem_avail_bytes{job="node-exporter",fstype!=""} / node_filesystem_size_bytes{job="node-exporter",fstype!=""} * 100 < 20
      and
        predict_linear(node_filesystem_avail_bytes{job="node-exporter",fstype!=""}[6h], 4*60*60) < 0
      and
        node_filesystem_readonly{job="node-exporter",fstype!=""} == 0
      )
    for: 1h
    labels:
      severity: critical
  - alert: 节点文件系统只剩下不到5%
    annotations:
      message: '{{ $labels.cluster }} 集群的 {{ $labels.instance }} 节点的文件系统的 {{ $labels.device }} 设备只剩下 {{ printf "%.2f" $value }}% 可使用空间！'
    expr: |
      (
        node_filesystem_avail_bytes{job="node-exporter",fstype!=""} / node_filesystem_size_bytes{job="node-exporter",fstype!=""} * 100 < 5
      and
        node_filesystem_readonly{job="node-exporter",fstype!=""} == 0
      )
    for: 1h
    labels:
      severity: warning
  - alert: 节点文件系统只剩下不到3%
    annotations:
      message: '{{ $labels.cluster }} 集群的 {{ $labels.instance }} 节点的文件系统的 {{ $labels.device }} 设备只剩下 {{ printf "%.2f" $value }}% 可使用空间！'
    expr: |
      (
        node_filesystem_avail_bytes{job="node-exporter",fstype!=""} / node_filesystem_size_bytes{job="node-exporter",fstype!=""} * 100 < 3
      and
        node_filesystem_readonly{job="node-exporter",fstype!=""} == 0
      )
    for: 1h
    labels:
      severity: critical
  - alert: 节点挂载的文件系统空闲的文件节点个数24小时内用完
    annotations:
      message: '{{ $labels.cluster }} 集群的 {{ $labels.instance }} 节点的文件系统的 {{ $labels.device }} 设备只剩下 {{ printf "%.2f" $value }}% 可使用空间，速率计算可能在24小时内填满！'
    expr: |
      (
        node_filesystem_files_free{job="node-exporter",fstype!=""} / node_filesystem_files{job="node-exporter",fstype!=""} * 100 < 40
      and
        predict_linear(node_filesystem_files_free{job="node-exporter",fstype!=""}[6h], 24*60*60) < 0
      and
        node_filesystem_readonly{job="node-exporter",fstype!=""} == 0
      )
    for: 1h
    labels:
      severity: warning
  - alert: 节点挂载的文件系统空闲的文件节点个数4小时内用完
    annotations:
      message: '{{ $labels.cluster }} 集群的 {{ $labels.instance }} 节点的文件系统的 {{ $labels.device }} 设备只剩下 {{ printf "%.2f" $value }}% 可使用空间，速率计算可能在4小时内填满！'
    expr: |
      (
        node_filesystem_files_free{job="node-exporter",fstype!=""} / node_filesystem_files{job="node-exporter",fstype!=""} * 100 < 20
      and
        predict_linear(node_filesystem_files_free{job="node-exporter",fstype!=""}[6h], 4*60*60) < 0
      and
        node_filesystem_readonly{job="node-exporter",fstype!=""} == 0
      )
    for: 1h
    labels:
      severity: critical
  - alert: 节点挂载的文件系统空闲的文件节点个数不到5%
    annotations:
      message: '{{ $labels.cluster }} 集群的 {{ $labels.instance }} 节点的文件系统的 {{ $labels.device }} 设备只剩下 {{ printf "%.2f" $value }}% 可使用空间！'
    expr: |
      (
        node_filesystem_files_free{job="node-exporter",fstype!=""} / node_filesystem_files{job="node-exporter",fstype!=""} * 100 < 5
      and
        node_filesystem_readonly{job="node-exporter",fstype!=""} == 0
      )
    for: 1h
    labels:
      severity: warning
  - alert: 节点挂载的文件系统空闲的文件节点个数不到3%
    annotations:
      message: '{{ $labels.cluster }} 集群的 {{ $labels.instance }} 节点的文件系统的 {{ $labels.device }} 设备只剩下 {{ printf "%.2f" $value }}% 可使用空间！'
    expr: |
      (
        node_filesystem_files_free{job="node-exporter",fstype!=""} / node_filesystem_files{job="node-exporter",fstype!=""} * 100 < 3
      and
        node_filesystem_readonly{job="node-exporter",fstype!=""} == 0
      )
    for: 1h
    labels:
      severity: critical
  - alert: 节点CPU使用大于85%
    expr: 100 - (avg by(cluster,instance) (irate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > 85
    for: 3m
    labels:
      severity: critical
    annotations:
      message: '{{ $labels.cluster }} 集群 {{ $labels.instance }} 节点 CPU 使用率在 3m 内持续达到 {{ printf "%.0f" $value }}%！'
  - alert: 节点内存空闲低
    annotations:
      message: '{{ $labels.cluster }} 集群 {{ $labels.instance }} 节点侦测到内存使用率在 3m 内持续达到 {{ printf "%.0f" $value }}%！'
    expr: |
      100 - ( node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes * 100 ) > 85
    for: 3m
    labels:
      severity: critical
  - alert: 侦测到OOM触发行为
    annotations:
      message: '{{ $labels.cluster }} 集群 {{ $labels.instance }} 节点侦测到 OOM 行为！'
    expr: |
      increase(node_vmstat_oom_kill[5m]) > 0
    for: 3m
    labels:
      severity: critical
#  - alert: 节点侦测到文件描述符切换次数过高
#    expr: (rate(node_context_switches_total[5m])) / (count without(cpu, mode) (node_cpu_seconds_total{mode="idle"})) > 5000
#    for: 1m
#    labels:
#      severity: critical
#    annotations:
#      message: '{{ $labels.cluster }} 集群 {{ $labels.instance }} 节点侦测到文件描述符切换次数达到 {{ printf "%.0f" $value }} 次/s!'
#  - alert: 节点侦测到打开的文件描述符过多
#    expr: avg by (instance,cluster) (node_filefd_allocated) > 102400
#    for: 1m
#    labels:
#      severity: critical
#    annotations:
#      message: '{{ $labels.cluster }} 集群 {{ $labels.instance }} 节点侦测到打开的文件描述符达到 {{ printf "%.0f" $value }}！'
```

## 14. time



```yaml
groups:
- name: node-time
  rules:
  - alert: 侦测到时钟偏差
    annotations:
      message:  '{{ $labels.cluster }} 集群 node-exporter {{ $labels.namespace }}/{{ $labels.pod }} 侦测到时钟偏差！'
    expr: |
      abs(node_timex_offset_seconds{job="node-exporter"}) > 0.05
    for: 2m
    labels:
      severity: warning
```



## 15. version

```yaml

```


