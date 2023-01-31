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

package controllers

import (
	"bytes"
	"context"
	"fmt"
	"time"

	cons "github.com/linkall-labs/vanus-operator/internal/constants"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	vanusv1alpha1 "github.com/linkall-labs/vanus-operator/api/v1alpha1"
)

// TriggerReconciler reconciles a Trigger object
type TriggerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=vanus.linkall.com,resources=triggers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=vanus.linkall.com,resources=triggers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=vanus.linkall.com,resources=triggers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Trigger object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *TriggerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	logger := log.Log.WithName("Trigger")
	logger.Info("Reconciling Trigger.")

	// Fetch the Trigger instance
	trigger := &vanusv1alpha1.Trigger{}
	err := r.Get(ctx, req.NamespacedName, trigger)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile req.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			logger.Info("Trigger resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the req.
		logger.Error(err, "Failed to get Trigger.")
		return ctrl.Result{RequeueAfter: time.Duration(cons.RequeueIntervalInSecond) * time.Second}, err
	}

	triggerDeployment := r.getDeploymentForTrigger(trigger)
	// Create Trigger Deployment
	// Check if the Deployment already exists, if not create a new one
	dep := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: triggerDeployment.Name, Namespace: triggerDeployment.Namespace}, dep)
	if err != nil {
		if errors.IsNotFound(err) {
			// Create Trigger ConfigMap
			triggerConfigMap := r.generateConfigMapForTrigger(trigger)
			logger.Info("Creating a new Trigger ConfigMap.", "ConfigMap.Namespace", triggerConfigMap.Namespace, "ConfigMap.Name", triggerConfigMap.Name)
			err = r.Create(ctx, triggerConfigMap)
			if err != nil {
				logger.Error(err, "Failed to create new Trigger ConfigMap", "ConfigMap.Namespace", triggerConfigMap.Namespace, "ConfigMap.Name", triggerConfigMap.Name)
				return ctrl.Result{}, err
			} else {
				logger.Info("Successfully create Trigger ConfigMap")
			}
			logger.Info("Creating a new Trigger Deployment.", "Deployment.Namespace", triggerDeployment.Namespace, "Deployment.Name", triggerDeployment.Name)
			err = r.Create(ctx, triggerDeployment)
			if err != nil {
				logger.Error(err, "Failed to create new Trigger Deployment", "Deployment.Namespace", triggerDeployment.Namespace, "Deployment.Name", triggerDeployment.Name)
				return ctrl.Result{}, err
			} else {
				logger.Info("Successfully create Trigger Deployment")
			}
			return ctrl.Result{}, nil
		} else {
			logger.Error(err, "Failed to get Trigger Deployment.")
			return ctrl.Result{RequeueAfter: time.Duration(cons.RequeueIntervalInSecond) * time.Second}, err
		}
	}

	// TODO(jiangkai): Update Trigger Deployment
	// TODO(jiangkai): Currently, only the updated mirror version is supported
	if trigger.Spec.Image != dep.Spec.Template.Spec.Containers[0].Image {
		logger.Info("Updating Trigger Deployment.", "Deployment.Namespace", triggerDeployment.Namespace, "Deployment.Name", triggerDeployment.Name)
		dep.Spec.Template.Spec.Containers[0].Image = trigger.Spec.Image
		err = r.Update(ctx, dep)
		if err != nil {
			logger.Error(err, "Failed to update Trigger Deployment", "Deployment.Namespace", triggerDeployment.Namespace, "Deployment.Name", triggerDeployment.Name)
			return ctrl.Result{}, err
		} else {
			logger.Info("Successfully update Trigger Deployment")
		}
	}
	if *trigger.Spec.Replicas != *dep.Spec.Replicas {
		logger.Info("Updating Trigger Deployment.", "Deployment.Namespace", triggerDeployment.Namespace, "Deployment.Name", triggerDeployment.Name)
		dep.Spec.Replicas = trigger.Spec.Replicas
		err = r.Update(ctx, dep)
		if err != nil {
			logger.Error(err, "Failed to update Trigger Deployment", "Deployment.Namespace", triggerDeployment.Namespace, "Deployment.Name", triggerDeployment.Name)
			return ctrl.Result{}, err
		} else {
			logger.Info("Successfully update Trigger Deployment")
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TriggerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&vanusv1alpha1.Trigger{}).
		Complete(r)
}

func (r *TriggerReconciler) generateConfigMapForTrigger(trigger *vanusv1alpha1.Trigger) *corev1.ConfigMap {
	data := make(map[string]string)
	value := bytes.Buffer{}
	value.WriteString("port: 2148\n")
	value.WriteString("ip: ${POD_IP}\n")
	value.WriteString("controllers:\n")
	// TODO(jiangkai): The timer needs to know the number of replicas of the controller，current default 3 replicas. Suggestted to use the service domain name for forwarding.
	for i := int32(0); i < 3; i++ {
		value.WriteString(fmt.Sprintf("  - vanus-controller-%d.vanus-controller.vanus.svc:2048\n", i))
	}
	data["trigger.yaml"] = value.String()
	triggerConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  trigger.Namespace,
			Name:       "config-trigger",
			Finalizers: []string{metav1.FinalizerOrphanDependents},
		},
		Data: data,
	}

	controllerutil.SetControllerReference(trigger, triggerConfigMap, r.Scheme)
	return triggerConfigMap
}

// returns a Trigger Deployment object
func (r *TriggerReconciler) getDeploymentForTrigger(trigger *vanusv1alpha1.Trigger) *appsv1.Deployment {
	labels := labelsForTrigger(trigger.Name)
	annotations := annotationsForTrigger()
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      trigger.Name,
			Namespace: trigger.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: trigger.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: cons.ServiceAccountName,
					Containers: []corev1.Container{{
						Name:            cons.TriggerContainerName,
						Image:           trigger.Spec.Image,
						ImagePullPolicy: trigger.Spec.ImagePullPolicy,
						Resources:       trigger.Spec.Resources,
						Env:             getEnvForTrigger(trigger),
						Ports:           getPortsForTrigger(trigger),
						VolumeMounts:    getVolumeMountsForTrigger(trigger),
					}},
					Volumes: getVolumesForTrigger(trigger),
				},
			},
		},
	}
	// Set Trigger instance as the owner and controller
	controllerutil.SetControllerReference(trigger, dep, r.Scheme)

	return dep
}

func getEnvForTrigger(trigger *vanusv1alpha1.Trigger) []corev1.EnvVar {
	defaultEnvs := []corev1.EnvVar{{
		Name:      cons.EnvPodIP,
		ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"}},
	}, {
		Name:      cons.EnvPodName,
		ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"}},
	}, {
		Name:  cons.EnvLogLevel,
		Value: "INFO",
	}}
	return defaultEnvs
}

func getPortsForTrigger(trigger *vanusv1alpha1.Trigger) []corev1.ContainerPort {
	defaultPorts := []corev1.ContainerPort{{
		Name:          cons.ContainerPortNameGrpc,
		ContainerPort: cons.TriggerPortGrpc,
	}}
	return defaultPorts
}

func getVolumeMountsForTrigger(trigger *vanusv1alpha1.Trigger) []corev1.VolumeMount {
	defaultVolumeMounts := []corev1.VolumeMount{{
		MountPath: cons.ConfigMountPath,
		Name:      cons.TriggerConfigMapName,
	}}
	return defaultVolumeMounts
}

func getVolumesForTrigger(trigger *vanusv1alpha1.Trigger) []corev1.Volume {
	defaultVolumes := []corev1.Volume{{
		Name: cons.TriggerConfigMapName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cons.TriggerConfigMapName,
				},
			}},
	}}
	return defaultVolumes
}

func labelsForTrigger(name string) map[string]string {
	return map[string]string{"app": name}
}

func annotationsForTrigger() map[string]string {
	return map[string]string{"vanus.dev/metrics.port": fmt.Sprintf("%d", cons.ControllerPortMetrics)}
}
