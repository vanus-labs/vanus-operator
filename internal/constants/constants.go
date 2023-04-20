// Copyright 2023 Linkall Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package constants defines some global constants
package constants

const (
	DefaultNamespace               = "vanus"
	DefaultVanusCoreName           = "vanus-cluster"
	DefaultVanusOperatorName       = "vanus-operator"
	DefaultVanusOperatorHostPrefix = "vanus-operator"
	DefaultConfigMountPath         = "/vanus/config"
	DefaultVolumeMountPath         = "/data"
	DefaultVolumeName              = "data"
	DefaultImagePullPolicy         = "Always"

	// DefaultRequeueIntervalInSecond is an universal interval of the reconcile function
	DefaultRequeueIntervalInSecond     = 6
	DefaultPortMetrics                 = 2112
	DefaultOperatorContainerPortApi    = 8089
	DefaultIngressClassAnnotationKey   = "kubernetes.io/ingress.class"
	DefaultIngressClassAnnotationValue = "alb"
)

const (
	// Controller
	DefaultControllerComponentName      = "vanus-controller"
	DefaultControllerContainerName      = "controller"
	DefaultControllerPortGrpc           = 2048
	DefaultControllerSegmentCapacity    = "64Mi" // 64Mi: 64*1024*1024=67108864
	DefaultControllerContainerImageName = "public.ecr.aws/vanus/controller"
	DefaultControllerConfigMapName      = "config-controller"
	// Root Controller
	DefaultRootControllerComponentName      = "vanus-root-controller"
	DefaultRootControllerContainerName      = "root-controller"
	DefaultRootControllerPortGrpc           = 2021
	DefaultRootControllerContainerImageName = "public.ecr.aws/vanus/root-controller"
	DefaultRootControllerConfigMapName      = "config-root-controller"
	// Etcd
	DefaultEtcdComponentName      = "vanus-etcd"
	DefaultEtcdContainerName      = "etcd"
	DefaultEtcdPortClient         = 2379
	DefaultEtcdPortPeer           = 2380
	DefaultEtcdContainerImageName = "public.ecr.aws/vanus/etcd:v3.5.7" // from 'docker.io/bitnami/etcd:3.5.7-debian-11-r9'
	DefaultEtcdVolumeMountPath    = "/bitnami/etcd"
	DefaultEtcdStorageSize        = "10Gi"
	// Store
	DefaultStoreComponentName      = "vanus-store"
	DefaultStoreContainerName      = "store"
	DefaultStoreContainerPortGrpc  = 11811
	DefaultStoreContainerImageName = "public.ecr.aws/vanus/store"
	DefaultStoreConfigMapName      = "config-store"
	DefaultStoreStorageSize        = "10Gi"
	// Trigger
	DefaultTriggerComponentName      = "vanus-trigger"
	DefaultTriggerContainerName      = "trigger"
	DefaultTriggerContainerPortGrpc  = 2148
	DefaultTriggerContainerImageName = "public.ecr.aws/vanus/trigger"
	DefaultTriggerConfigMapName      = "config-trigger"
	// Timer
	DefaultTimerComponentName      = "vanus-timer"
	DefaultTimerContainerName      = "timer"
	DefaultTimerContainerImageName = "public.ecr.aws/vanus/timer"
	DefaultTimerConfigMapName      = "config-timer"
	DefaultTimerTimingWheelTick    = 1
	DefaultTimerTimingWheelSize    = 32
	DefaultTimerTimingWheelLayers  = 4
	// Gateway
	DefaultGatewayComponentName              = "vanus-gateway"
	DefaultGatewayContainerName              = "gateway"
	DefaultGatewayContainerPortProxy         = 8080
	DefaultGatewayContainerPortCloudevents   = 8081
	DefaultGatewayContainerPortSinkProxy     = 8082
	DefaultGatewayServiceNodePortProxy       = 30001
	DefaultGatewayServiceNodePortCloudevents = 30002
	DefaultGatewayContainerImageName         = "public.ecr.aws/vanus/gateway"
	DefaultGatewayConfigMapName              = "config-gateway"
	// Connector
	DefaultConnectorContainerName   = "connector"
	DefaultConnectorConfigMountPath = "/vanus-connect/config"
	DefaultConnectorConfigMapName   = "config"
	DefaultConnectorServiceType     = "ClusterIP"
	DefaultConnectorServicePort     = 80
)

const (
	// OperatorServiceAccountName is the ServiceAccount name of Vanus cluster
	OperatorServiceAccountName = "vanus-operator"
	HeadlessServiceClusterIP   = "None"

	ContainerPortNameGrpc        = "grpc"
	ContainerPortNameClient      = "client"
	ContainerPortNamePeer        = "peer"
	ContainerPortNameMetrics     = "metrics"
	ContainerPortNameProxy       = "proxy"
	ContainerPortNameCloudevents = "cloudevents"
	ContainerPortNameSinkProxy   = "sinkproxy"

	EnvPodIP    = "POD_IP"
	EnvPodName  = "POD_NAME"
	EnvLogLevel = "VANUS_LOG_LEVEL"

	AnnotationBuildInIngress = "vanus.ai/build-in-ingress"
)

var (
	DefaultControllerReplicas int32 = 2
	DefaultEtcdReplicas       int32 = 3
	DefaultStoreReplicas      int32 = 3
	DefaultGatewayReplicas    int32 = 1
	DefaultTriggerReplicas    int32 = 1
	DefaultTimerReplicas      int32 = 2
	DefaultConnectorReplicas  int32 = 1
)

// Annotations supported by Core
const (
	CoreComponentImagePullPolicyAnnotation = "core.vanus.ai/image-pull-policy"
	// Etcd
	CoreComponentEtcdPortClientAnnotation        = "core.vanus.ai/etcd-port-client"
	CoreComponentEtcdPortPeerAnnotation          = "core.vanus.ai/etcd-port-peer"
	CoreComponentEtcdReplicasAnnotation          = "core.vanus.ai/etcd-replicas"
	CoreComponentEtcdStorageSizeAnnotation       = "core.vanus.ai/etcd-storage-size"
	CoreComponentEtcdStorageClassAnnotation      = "core.vanus.ai/etcd-storage-class"
	CoreComponentEtcdResourceLimitsCpuAnnotation = "core.vanus.ai/etcd-resource-limits-cpu"
	CoreComponentEtcdResourceLimitsMemAnnotation = "core.vanus.ai/etcd-resource-limits-mem"
	// Controller
	CoreComponentControllerSvcPortAnnotation           = "core.vanus.ai/controller-service-port"
	CoreComponentControllerReplicasAnnotation          = "core.vanus.ai/controller-replicas"
	CoreComponentControllerSegmentCapacityAnnotation   = "core.vanus.ai/controller-segment-capacity"
	CoreComponentControllerResourceLimitsCpuAnnotation = "core.vanus.ai/controller-resource-limits-cpu"
	CoreComponentControllerResourceLimitsMemAnnotation = "core.vanus.ai/controller-resource-limits-mem"
	// Root Controller
	CoreComponentRootControllerSvcPortAnnotation = "core.vanus.ai/root-controller-service-port"
	// Store
	CoreComponentStoreReplicasAnnotation          = "core.vanus.ai/store-replicas"
	CoreComponentStoreStorageSizeAnnotation       = "core.vanus.ai/store-storage-size"
	CoreComponentStoreStorageClassAnnotation      = "core.vanus.ai/store-storage-class"
	CoreComponentStoreResourceLimitsCpuAnnotation = "core.vanus.ai/store-resource-limits-cpu"
	CoreComponentStoreResourceLimitsMemAnnotation = "core.vanus.ai/store-resource-limits-mem"
	// Gateway
	CoreComponentGatewayPortProxyAnnotation           = "core.vanus.ai/gateway-port-proxy"
	CoreComponentGatewayPortCloudEventsAnnotation     = "core.vanus.ai/gateway-port-cloudevents"
	CoreComponentGatewayNodePortProxyAnnotation       = "core.vanus.ai/gateway-nodeport-proxy"
	CoreComponentGatewayNodePortCloudEventsAnnotation = "core.vanus.ai/gateway-nodeport-cloudevents"
	CoreComponentGatewayReplicasAnnotation            = "core.vanus.ai/gateway-replicas"
	CoreComponentGatewayResourceLimitsCpuAnnotation   = "core.vanus.ai/gateway-resource-limits-cpu"
	CoreComponentGatewayResourceLimitsMemAnnotation   = "core.vanus.ai/gateway-resource-limits-mem"
	// Trigger
	CoreComponentTriggerReplicasAnnotation          = "core.vanus.ai/trigger-replicas"
	CoreComponentTriggerResourceLimitsCpuAnnotation = "core.vanus.ai/trigger-resource-limits-cpu"
	CoreComponentTriggerResourceLimitsMemAnnotation = "core.vanus.ai/trigger-resource-limits-mem"
	// Timer
	CoreComponentTimerReplicasAnnotation          = "core.vanus.ai/timer-replicas"
	CoreComponentTimerTimingWheelTickAnnotation   = "core.vanus.ai/timer-timingwheel-tick"
	CoreComponentTimerTimingWheelSizeAnnotation   = "core.vanus.ai/timer-timingwheel-size"
	CoreComponentTimerTimingWheelLayersAnnotation = "core.vanus.ai/timer-timingwheel-layers"
	CoreComponentTimerResourceLimitsCpuAnnotation = "core.vanus.ai/timer-resource-limits-cpu"
	CoreComponentTimerResourceLimitsMemAnnotation = "core.vanus.ai/timer-resource-limits-mem"
)

// Annotations supported by Connector
const (
	ConnectorDeploymentReplicasAnnotation = "connector.vanus.ai/deployment-replicas"
	ConnectorServiceTypeAnnotation        = "connector.vanus.ai/service-type"
	ConnectorServicePortAnnotation        = "connector.vanus.ai/service-port"
	ConnectorNetworkHostDomainAnnotation  = "connector.vanus.ai/network-host-domain"
	ConnectorIngressNameAnnotation        = "connector.vanus.ai/ingress-name"
	ConnectorRestartAtAnnotation          = "connector.vanus.ai/restart-at"
)

const (
	// PodStatusStatusPending captures enum value "Pending"
	PodStatusStatusPending string = "Pending"

	// PodStatusStatusRunning captures enum value "Running"
	PodStatusStatusRunning string = "Running"

	// PodStatusStatusSucceeded captures enum value "Succeeded"
	PodStatusStatusSucceeded string = "Succeeded"

	// PodStatusStatusStarting captures enum value "Starting"
	PodStatusStatusStarting string = "Starting"

	// PodStatusStatusFailed captures enum value "Failed"
	PodStatusStatusFailed string = "Failed"

	// PodStatusStatusRemoving captures enum value "Removing"
	PodStatusStatusRemoving string = "Removing"

	// PodStatusStatusUnknown captures enum value "Unknown"
	PodStatusStatusUnknown string = "Unknown"
)
