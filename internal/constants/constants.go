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
	DefaultNamespace        = "vanus"
	DefaultVanusClusterName = "vanus-cluster"

	DefaultOperatorName       = "vanus-operator"
	DefaultControllerName     = "vanus-controller"
	DefaultRootControllerName = "vanus-root-controller"
	DefaultEtcdName           = "vanus-etcd"
	DefaultStoreName          = "vanus-store"
	DefaultTriggerName        = "vanus-trigger"
	DefaultTimerName          = "vanus-timer"
	DefaultGatewayName        = "vanus-gateway"

	DefaultEtcdStorageSize  = "10Gi"
	DefaultStoreStorageSize = "10Gi"
)

const (
	CoreComponentEtcdReplicasAnnotation       = "core.vanus.ai/etcd-replicas"
	CoreComponentControllerReplicasAnnotation = "core.vanus.ai/controller-replicas"
	CoreComponentStoreReplicasAnnotation      = "core.vanus.ai/store-replicas"
	CoreComponentGatewayReplicasAnnotation    = "core.vanus.ai/gateway-replicas"
	CoreComponentTriggerReplicasAnnotation    = "core.vanus.ai/trigger-replicas"
	CoreComponentTimerReplicasAnnotation      = "core.vanus.ai/timer-replicas"
	CoreComponentEtcdStorageSizeAnnotation    = "core.vanus.ai/etcd-storage-size"
	CoreComponentEtcdStorageClassAnnotation   = "core.vanus.ai/etcd-storage-class"
	CoreComponentStoreStorageSizeAnnotation   = "core.vanus.ai/store-storage-size"
	CoreComponentStoreStorageClassAnnotation  = "core.vanus.ai/store-storage-class"
	ConnectorServiceTypeAnnotation            = "connector.vanus.ai/service-type"
	ConnectorServicePortAnnotation            = "connector.vanus.ai/service-port"
	ConnectorNetworkHostDomainAnnotation      = "connector.vanus.ai/network-host-domain"
)

const (
	// ControllerContainerName is the name of Controller container
	ControllerContainerName = "controller"

	// ControllerImageName is the name of Controller container image
	ControllerImageName = "public.ecr.aws/vanus/controller"

	// ControllerConfigMapName is the name of Controller configmap
	ControllerConfigMapName = "config-controller"

	// RootControllerContainerName is the name of Controller container
	RootControllerContainerName = "root-controller"

	// ControllerImageName is the name of Controller container image
	RootControllerImageName = "public.ecr.aws/vanus/root-controller"

	// ControllerConfigMapName is the name of Controller configmap
	RootControllerConfigMapName = "config-root-controller"

	// EtcdInitContainerName is the name of init Controller container
	EtcdInitContainerName = "init"

	// EtcdContainerName is the name of Controller container
	EtcdContainerName = "etcd"

	// EtcdInitContainerImageName is the name of init Controller container image, from 'docker.io/bitnami/etcd:3.5.7-debian-11-r9'
	EtcdInitContainerImageName = "public.ecr.aws/vanus/metadata-migration:latest"

	// EtcdImageName is the name of Controller container image, from 'docker.io/bitnami/etcd:3.5.7-debian-11-r9'
	EtcdImageName = "public.ecr.aws/vanus/etcd:v3.5.7"

	// EtcdInitContainerVolumeMountName is the name of Store volume mount
	EtcdInitContainerVolumeMountName = "init-snapshot"

	// EtcdInitContainerVolumeMountPath is the directory of Store data files
	EtcdInitContainerVolumeMountPath = "/init-snapshot"

	// EtcdVolumeMountPath is the directory of Store data files
	EtcdVolumeMountPath = "/bitnami/etcd"

	// StoreContainerName is the name of store container
	StoreContainerName = "store"

	// StoreImageName is the name of Store container image
	StoreImageName = "public.ecr.aws/vanus/store"

	// StoreConfigMapName is the name of mounted configuration file
	StoreConfigMapName = "config-store"

	// TriggerContainerName is the name of trigger container
	TriggerContainerName = "trigger"

	// TriggerImageName is the name of Trigger container image
	TriggerImageName = "public.ecr.aws/vanus/trigger"

	// TriggerConfigMapName is the name of mounted configuration file
	TriggerConfigMapName = "config-trigger"

	// TimerContainerName is the name of timer container
	TimerContainerName = "timer"

	// TimerImageName is the name of Timer container image
	TimerImageName = "public.ecr.aws/vanus/timer"

	// TimerConfigMapName is the name of mounted configuration file
	TimerConfigMapName = "config-timer"

	// GatewayContainerName is the name of gateway container
	GatewayContainerName = "gateway"

	// GatewayImageName is the name of Gateway container image
	GatewayImageName = "public.ecr.aws/vanus/gateway"

	// GatewayConfigMapName is the name of mounted configuration file
	GatewayConfigMapName = "config-gateway"

	// ConnectorContainerName is the name of Connector container
	ConnectorContainerName = "connector"

	// ConnectorConfigMapName is the name of Connector configmap
	ConnectorConfigMapName = "config-connector"

	// ConfigMountPath is the directory of Vanus configd files
	ConfigMountPath = "/vanus/config"

	// VanceConfigMountPath is the directory of Vance configd files
	VanceConfigMountPath = "/vanus-connect/config"

	// VanceConfigMapName is the directory of Vance configd files
	VanceConfigMapName = "config"

	// VolumeMountPath is the directory of Store data files
	VolumeMountPath = "/data"

	// VolumeName is the directory of Store data files
	VolumeName = "data"

	// VolumeStorage is the directory of Store data files
	VolumeStorage = "10Gi"

	// StorageModeStorageClass is the name of StorageClass storage mode
	StorageModeStorageClass = "StorageClass"

	// StorageModeEmptyDir is the name of EmptyDir storage mode
	StorageModeEmptyDir = "EmptyDir"

	// StorageModeHostPath is the name pf HostPath storage mode
	StorageModeHostPath = "HostPath"

	// EnvPodIP the container environment variable name of controller pod ip
	EnvPodIP = "POD_IP"

	// EnvPodName the container environment variable name of controller pod name
	EnvPodName = "POD_NAME"

	// EnvLogLevel the container environment variable name of controller log level
	EnvLogLevel = "VANUS_LOG_LEVEL"
)

const (
	// ServiceAccountName is the ServiceAccount name of Vanus cluster
	ServiceAccountName = "vanus-operator"

	ContainerPortNameGrpc        = "grpc"
	ContainerPortNameEtcdClient  = "etcd-client"
	ContainerPortNameEtcdPeer    = "etcd-peer"
	ContainerPortNameMetrics     = "metrics"
	ContainerPortNameProxy       = "proxy"
	ContainerPortNameCloudevents = "cloudevents"
	ContainerPortNameSinkProxy   = "sinkproxy"

	ControllerPortGrpc    = 2048
	ControllerPortMetrics = 2112

	RootControllerPortGrpc    = 2021
	RootControllerPortMetrics = 2112

	EtcdPortClient = 2379
	EtcdPortPeer   = 2380

	StorePortGrpc              = 11811
	TriggerPortGrpc            = 2148
	GatewayNodePortProxy       = 30001
	GatewayNodePortCloudevents = 30002
	GatewayPortProxy           = 8080
	GatewayPortCloudevents     = 8081
	GatewayPortSinkProxy       = 8082

	HeadlessService = "None"

	// RequeueIntervalInSecond is an universal interval of the reconcile function
	RequeueIntervalInSecond = 6
)

var (
	// GroupNum is the number of broker group
	GroupNum = 0

	// NameServersStr is the name server list
	NameServersStr = ""

	// IsNameServersStrUpdated is whether the name server list is updated
	IsNameServersStrUpdated = false

	// IsNameServersStrInitialized is whether the name server list is initialized
	IsNameServersStrInitialized = false

	// BrokerClusterName is the broker cluster name
	BrokerClusterName = ""

	// svc of controller for brokers
	ControllerAccessPoint = ""
)

var (
	DefaultControllerReplicas int32 = 2
	DefaultEtcdReplicas       int32 = 3
	DefaultStoreReplicas      int32 = 3
	DefaultGatewayReplicas    int32 = 1
	DefaultTriggerReplicas    int32 = 1
	DefaultTimerReplicas      int32 = 2
)
