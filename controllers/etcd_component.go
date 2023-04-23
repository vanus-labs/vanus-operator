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

package controllers

import (
	"context"
	stderr "errors"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	cons "github.com/vanus-labs/vanus-operator/pkg/constants"
	"github.com/vanus-labs/vanus-operator/pkg/convert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	vanusv1alpha1 "github.com/vanus-labs/vanus-operator/api/v1alpha1"
)

func (r *CoreReconciler) handleEtcd(ctx context.Context, logger logr.Logger, core *vanusv1alpha1.Core) (ctrl.Result, error) {
	// Create Etcd StatefulSet
	// Check if the statefulSet already exists, if not create a new one
	sts := &appsv1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{Name: cons.DefaultEtcdComponentName, Namespace: cons.DefaultNamespace}, sts)
	if err != nil {
		if errors.IsNotFound(err) {
			// Create Etcd Service
			etcd := r.generateEtcd(core)
			logger.Info("Creating a new Etcd StatefulSet.", "Namespace", etcd.Namespace, "Name", etcd.Name)
			err = r.Create(ctx, etcd)
			if err != nil {
				logger.Error(err, "Failed to create new Etcd StatefulSet", "Namespace", etcd.Namespace, "Name", etcd.Name)
				return ctrl.Result{}, err
			} else {
				logger.Info("Successfully create Etcd StatefulSet")
			}
			etcdSvc := r.generateSvcForEtcd(core)
			// Create Etcd Service
			// Check if the service already exists, if not create a new one
			svc := &corev1.Service{}
			err = r.Get(ctx, types.NamespacedName{Name: etcdSvc.Name, Namespace: etcdSvc.Namespace}, svc)
			if err != nil {
				if errors.IsNotFound(err) {
					logger.Info("Creating a new Etcd Service.", "Namespace", etcdSvc.Namespace, "Name", etcdSvc.Name)
					err = r.Create(ctx, etcdSvc)
					if err != nil {
						logger.Error(err, "Failed to create new Etcd Service", "Namespace", etcdSvc.Namespace, "Name", etcdSvc.Name)
						return ctrl.Result{}, err
					} else {
						logger.Info("Successfully create Etcd Service")
					}
				} else {
					logger.Error(err, "Failed to get Etcd Service.")
					return ctrl.Result{RequeueAfter: time.Duration(cons.DefaultRequeueIntervalInSecond) * time.Second}, err
				}
			}

			// Wait for Etcd is ready
			start := time.Now()
			logger.Info("Wait for Etcd install is ready")
			ticker := time.NewTicker(defaultWaitForReadyTimeout)
			defer ticker.Stop()
			for {
				ready, err := r.waitEtcdIsReady(ctx)
				if err != nil {
					logger.Error(err, "Wait for Etcd install is ready but got error")
					return ctrl.Result{RequeueAfter: time.Duration(cons.DefaultRequeueIntervalInSecond) * time.Second}, err
				}
				if ready {
					break
				}
				select {
				case <-ticker.C:
					return ctrl.Result{RequeueAfter: time.Duration(cons.DefaultRequeueIntervalInSecond) * time.Second}, stderr.New("etcd cluster isn't ready")
				default:
					time.Sleep(time.Second)
				}
			}
			logger.Info("Etcd is ready", "WaitingTime", time.Since(start))
			return ctrl.Result{}, nil
		} else {
			logger.Error(err, "Failed to get Etcd StatefulSet.")
			return ctrl.Result{RequeueAfter: time.Duration(cons.DefaultRequeueIntervalInSecond) * time.Second}, err
		}
	}

	// Update Etcd StatefulSet Not Supported.

	return ctrl.Result{}, nil
}

func (r *CoreReconciler) waitEtcdIsReady(ctx context.Context) (bool, error) {
	sts := &appsv1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{Name: cons.DefaultEtcdComponentName, Namespace: cons.DefaultNamespace}, sts)
	if err != nil {
		return false, err
	}
	if sts.Status.Replicas == cons.DefaultEtcdReplicas && sts.Status.ReadyReplicas == cons.DefaultEtcdReplicas && sts.Status.AvailableReplicas == cons.DefaultEtcdReplicas && sts.Status.UpdatedReplicas == cons.DefaultEtcdReplicas {
		return true, nil
	}
	return false, nil
}

// returns a Etcd StatefulSet object
func (r *CoreReconciler) generateEtcd(core *vanusv1alpha1.Core) *appsv1.StatefulSet {
	var (
		allowPrivilegeEscalation = false
		runAsNonRoot             = true
		runAsUser                = int64(1001)
		fsGroup                  = int64(1001)
	)
	labels := genLabels(cons.DefaultEtcdComponentName)
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cons.DefaultEtcdComponentName,
			Namespace: cons.DefaultNamespace,
			Labels:    labels,
		},
		Spec: appsv1.StatefulSetSpec{
			PodManagementPolicy: appsv1.ParallelPodManagement,
			Replicas:            &cons.DefaultEtcdReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
			ServiceName: cons.DefaultEtcdComponentName,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:            cons.DefaultEtcdContainerName,
						Image:           cons.DefaultEtcdContainerImageName,
						ImagePullPolicy: corev1.PullPolicy(core.Annotations[cons.CoreComponentImagePullPolicyAnnotation]),
						Env:             getEnvForEtcd(core),
						Lifecycle: &corev1.Lifecycle{
							PreStop: &corev1.LifecycleHandler{
								Exec: &corev1.ExecAction{
									Command: []string{"/opt/bitnami/scripts/etcd/prestop.sh"},
								},
							},
						},
						LivenessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								Exec: &corev1.ExecAction{
									Command: []string{"/opt/bitnami/scripts/etcd/healthcheck.sh"},
								},
							},
							FailureThreshold:    5,
							InitialDelaySeconds: 60,
							PeriodSeconds:       30,
							SuccessThreshold:    1,
							TimeoutSeconds:      5,
						},
						ReadinessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								Exec: &corev1.ExecAction{
									Command: []string{"/opt/bitnami/scripts/etcd/healthcheck.sh"},
								},
							},
							FailureThreshold:    5,
							InitialDelaySeconds: 60,
							PeriodSeconds:       10,
							SuccessThreshold:    1,
							TimeoutSeconds:      5,
						},
						SecurityContext: &corev1.SecurityContext{
							AllowPrivilegeEscalation: &allowPrivilegeEscalation,
							RunAsNonRoot:             &runAsNonRoot,
							RunAsUser:                &runAsUser,
						},
						TerminationMessagePath:   "/dev/termination-log",
						TerminationMessagePolicy: corev1.TerminationMessageReadFile,
						Resources:                getResourcesForEtcd(core),
						Ports:                    getPortsForEtcd(core),
						VolumeMounts:             getVolumeMountsForEtcd(core),
					}},
					SecurityContext: &corev1.PodSecurityContext{
						FSGroup: &fsGroup,
					},
				},
			},
			VolumeClaimTemplates: getVolumeClaimTemplatesForEtcd(core),
		},
	}

	// Set Etcd instance as the owner and controller
	controllerutil.SetControllerReference(core, sts, r.Scheme)

	return sts
}

func getEnvForEtcd(core *vanusv1alpha1.Core) []corev1.EnvVar {
	defaultEnvs := []corev1.EnvVar{{
		Name:  "BITNAMI_DEBUG",
		Value: "false",
	}, {
		Name:      "MY_POD_IP",
		ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"}},
	}, {
		Name:      "MY_POD_NAME",
		ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"}},
	}, {
		Name:  "MY_STS_NAME",
		Value: "vanus-etcd",
	}, {
		Name:  "ETCDCTL_API",
		Value: "3",
	}, {
		Name:  "ETCD_ON_K8S",
		Value: "yes",
	}, {
		Name:  "ETCD_INIT_SNAPSHOT_FILENAME",
		Value: "snapshotdb",
	}, {
		Name:  "ETCD_DISASTER_RECOVERY",
		Value: "no",
	}, {
		Name:  "ETCD_NAME",
		Value: "$(MY_POD_NAME)",
	}, {
		Name:  "ETCD_DATA_DIR",
		Value: "/bitnami/etcd/data",
	}, {
		Name:  "ETCD_LOG_LEVEL",
		Value: "info",
	}, {
		Name:  "ALLOW_NONE_AUTHENTICATION",
		Value: "yes",
	}, {
		Name:  "ETCD_ADVERTISE_CLIENT_URLS",
		Value: fmt.Sprintf("http://$(MY_POD_NAME).vanus-etcd.vanus.svc.cluster.local:%s,http://vanus-etcd.vanus.svc.cluster.local:%s", core.Annotations[cons.CoreComponentEtcdPortClientAnnotation], core.Annotations[cons.CoreComponentEtcdPortClientAnnotation]),
	}, {
		Name:  "ETCD_LISTEN_CLIENT_URLS",
		Value: fmt.Sprintf("http://0.0.0.0:%s", core.Annotations[cons.CoreComponentEtcdPortClientAnnotation]),
	}, {
		Name:  "ETCD_INITIAL_ADVERTISE_PEER_URLS",
		Value: fmt.Sprintf("http://$(MY_POD_NAME).vanus-etcd.vanus.svc.cluster.local:%s", core.Annotations[cons.CoreComponentEtcdPortPeerAnnotation]),
	}, {
		Name:  "ETCD_LISTEN_PEER_URLS",
		Value: fmt.Sprintf("http://0.0.0.0:%s", core.Annotations[cons.CoreComponentEtcdPortPeerAnnotation]),
	}, {
		Name:  "ETCD_INITIAL_CLUSTER_STATE",
		Value: "new",
	}, {
		Name:  "ETCD_INITIAL_CLUSTER",
		Value: fmt.Sprintf("vanus-etcd-0=http://vanus-etcd-0.vanus-etcd.vanus.svc.cluster.local:%s,vanus-etcd-1=http://vanus-etcd-1.vanus-etcd.vanus.svc.cluster.local:%s,vanus-etcd-2=http://vanus-etcd-2.vanus-etcd.vanus.svc.cluster.local:%s", core.Annotations[cons.CoreComponentEtcdPortPeerAnnotation], core.Annotations[cons.CoreComponentEtcdPortPeerAnnotation], core.Annotations[cons.CoreComponentEtcdPortPeerAnnotation]),
	}, {
		Name:  "ETCD_CLUSTER_DOMAIN",
		Value: "vanus-etcd.vanus.svc.cluster.local",
	}}

	return defaultEnvs
}

func getResourcesForEtcd(core *vanusv1alpha1.Core) corev1.ResourceRequirements {
	limits := make(map[corev1.ResourceName]resource.Quantity)
	if val, ok := core.Annotations[cons.CoreComponentEtcdResourceLimitsCpuAnnotation]; ok && val != "" {
		limits[corev1.ResourceCPU] = resource.MustParse(core.Annotations[cons.CoreComponentEtcdResourceLimitsCpuAnnotation])
	}
	if val, ok := core.Annotations[cons.CoreComponentEtcdResourceLimitsMemAnnotation]; ok && val != "" {
		limits[corev1.ResourceMemory] = resource.MustParse(core.Annotations[cons.CoreComponentEtcdResourceLimitsMemAnnotation])
	}
	defaultResources := corev1.ResourceRequirements{
		Limits: limits,
	}
	return defaultResources
}

func getPortsForEtcd(core *vanusv1alpha1.Core) []corev1.ContainerPort {
	portClient, _ := convert.StrToInt32(core.Annotations[cons.CoreComponentEtcdPortClientAnnotation])
	portPeer, _ := convert.StrToInt32(core.Annotations[cons.CoreComponentEtcdPortPeerAnnotation])
	defaultPorts := []corev1.ContainerPort{{
		Name:          cons.ContainerPortNameClient,
		ContainerPort: portClient,
	}, {
		Name:          cons.ContainerPortNamePeer,
		ContainerPort: portPeer,
	}}
	return defaultPorts
}

func getVolumeMountsForEtcd(core *vanusv1alpha1.Core) []corev1.VolumeMount {
	defaultVolumeMounts := []corev1.VolumeMount{{
		MountPath: cons.DefaultEtcdVolumeMountPath,
		Name:      cons.DefaultVolumeName,
	}}
	return defaultVolumeMounts
}

func getVolumeClaimTemplatesForEtcd(core *vanusv1alpha1.Core) []corev1.PersistentVolumeClaim {
	labels := genLabels(cons.DefaultEtcdComponentName)
	requests := make(map[corev1.ResourceName]resource.Quantity)
	requests[corev1.ResourceStorage] = resource.MustParse(core.Annotations[cons.CoreComponentEtcdStorageSizeAnnotation])
	defaultPersistentVolumeClaims := []corev1.PersistentVolumeClaim{{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
			Name:   cons.DefaultVolumeName,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: requests,
			},
		},
	}}
	if val, ok := core.Annotations[cons.CoreComponentEtcdStorageClassAnnotation]; ok && val != "" {
		defaultPersistentVolumeClaims[0].Spec.StorageClassName = &val
	}
	return defaultPersistentVolumeClaims
}

func (r *CoreReconciler) generateSvcForEtcd(core *vanusv1alpha1.Core) *corev1.Service {
	portClient, _ := convert.StrToInt32(core.Annotations[cons.CoreComponentEtcdPortClientAnnotation])
	portPeer, _ := convert.StrToInt32(core.Annotations[cons.CoreComponentEtcdPortPeerAnnotation])
	labels := genLabels(cons.DefaultEtcdComponentName)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  cons.DefaultNamespace,
			Name:       cons.DefaultEtcdComponentName,
			Labels:     labels,
			Finalizers: []string{metav1.FinalizerOrphanDependents},
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: cons.HeadlessServiceClusterIP,
			Selector:  labels,
			Ports: []corev1.ServicePort{
				{
					Name:       cons.ContainerPortNameClient,
					Port:       portClient,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt(int(portClient)),
				}, {
					Name:       cons.ContainerPortNamePeer,
					Port:       portPeer,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromInt(int(portPeer)),
				}},
			PublishNotReadyAddresses: true,
		},
	}

	controllerutil.SetControllerReference(core, svc, r.Scheme)
	return svc
}
