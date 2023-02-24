// Copyright 2022 Linkall Inc.
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
	// ControllerContainerName is the name of Controller container
	ControllerContainerName = "controller"

	// ControllerConfigMapName is the name of Controller configmap
	ControllerConfigMapName = "config-controller"

	// StoreContainerName is the name of store container
	StoreContainerName = "store"

	// StoreConfigMapName is the name of mounted configuration file
	StoreConfigMapName = "config-store"

	// TriggerContainerName is the name of tigger container
	TriggerContainerName = "tigger"

	// TriggerConfigMapName is the name of mounted configuration file
	TriggerConfigMapName = "config-trigger"

	// TimerContainerName is the name of timer container
	TimerContainerName = "timer"

	// TimerConfigMapName is the name of mounted configuration file
	TimerConfigMapName = "config-timer"

	// GatewayContainerName is the name of gateway container
	GatewayContainerName = "gateway"

	// GatewayConfigMapName is the name of mounted configuration file
	GatewayConfigMapName = "config-gateway"

	// ConnectorContainerName is the name of Connector container
	ConnectorContainerName = "connector"

	// ConnectorConfigMapName is the name of Connector configmap
	ConnectorConfigMapName = "config-connector"

	// ConfigMountPath is the directory of Vanus config files
	ConfigMountPath = "/vanus/config"

	// VanusConnectConfigMountPath is the directory of vanus-connect config files
	VanusConnectConfigMountPath = "/vanus-connect/config"

	// VanusConnectConfigMapName is the directory of vanus-connect config files
	VanusConnectConfigMapName = "config"

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

	ControllerPortGrpc       = 2048
	ControllerPortEtcdClient = 2379
	ControllerPortEtcdPeer   = 2380
	ControllerPortMetrics    = 2112

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

const (
	DefaultNamespace = "vanus"

	DefaultControllerName = "vanus-controller"
	DefaultStoreName      = "vanus-store"
	DefaultTriggerName    = "vanus-trigger"
	DefaultTimerName      = "vanus-timer"
	DefaultGatewayName    = "vanus-gateway"
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

	// ControllerAccessPoint svc of controller for brokers
	ControllerAccessPoint = ""
)

var (
	DefaultControllerReplicas int32 = 3
	DefaultStoreReplicas      int32 = 3
)
