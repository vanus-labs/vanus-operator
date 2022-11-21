/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
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

// TimerReconciler reconciles a Timer object
type TimerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=vanus.linkall.com,resources=timers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=vanus.linkall.com,resources=timers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=vanus.linkall.com,resources=timers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Timer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *TimerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	logger := log.Log.WithName("Timer")
	logger.Info("Reconciling Timer.")

	// Fetch the Timer instance
	timer := &vanusv1alpha1.Timer{}
	err := r.Get(ctx, req.NamespacedName, timer)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile req.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			logger.Info("Timer resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the req.
		logger.Error(err, "Failed to get Timer.")
		return ctrl.Result{RequeueAfter: time.Duration(cons.RequeueIntervalInSecond) * time.Second}, err
	}

	timerDeployment := r.getDeploymentForTimer(timer)
	// Create Timer Deployment
	// Check if the Deployment already exists, if not create a new one
	dep := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: timerDeployment.Name, Namespace: timerDeployment.Namespace}, dep)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Creating a new Timer Deployment.", "Deployment.Namespace", timerDeployment.Namespace, "Deployment.Name", timerDeployment.Name)
			err = r.Create(ctx, timerDeployment)
			if err != nil {
				logger.Error(err, "Failed to create new Timer Deployment", "Deployment.Namespace", timerDeployment.Namespace, "Deployment.Name", timerDeployment.Name)
				return ctrl.Result{}, err
			} else {
				logger.Info("Successfully create Timer Deployment")
			}
		} else {
			logger.Error(err, "Failed to get Timer Deployment.")
			return ctrl.Result{RequeueAfter: time.Duration(cons.RequeueIntervalInSecond) * time.Second}, err
		}
	}

	// TODO(jiangkai): Update Timer Deployment

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TimerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&vanusv1alpha1.Timer{}).
		Complete(r)
}

// returns a Timer Deployment object
func (r *TimerReconciler) getDeploymentForTimer(timer *vanusv1alpha1.Timer) *appsv1.Deployment {
	labels := labelsForTimer(timer.Name)
	annotations := annotationsForTimer()
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      timer.Name,
			Namespace: timer.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: timer.Spec.Replicas,
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
						Resources:       timer.Spec.Resources,
						Image:           timer.Spec.Image,
						Name:            cons.TimerContainerName,
						ImagePullPolicy: timer.Spec.ImagePullPolicy,
						Env:             getEnvForTimer(timer),
						VolumeMounts:    getVolumeMountsForTimer(timer),
					}},
					Volumes: getVolumesForTimer(timer),
				},
			},
		},
	}
	// Set Timer instance as the owner and controller
	controllerutil.SetControllerReference(timer, dep, r.Scheme)

	return dep
}

func getEnvForTimer(timer *vanusv1alpha1.Timer) []corev1.EnvVar {
	defaultEnvs := []corev1.EnvVar{{
		Name:      cons.EnvPodIP,
		ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"}},
	}, {
		Name:      cons.EnvPodName,
		ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"}},
	}, {
		Name:  cons.EnvLogLevel,
		Value: "DEBUG",
	}}
	return defaultEnvs
}

func getVolumeMountsForTimer(timer *vanusv1alpha1.Timer) []corev1.VolumeMount {
	defaultVolumeMounts := []corev1.VolumeMount{{
		MountPath: cons.ConfigMountPath,
		Name:      cons.TimerConfigMapName,
	}}
	return defaultVolumeMounts
}

func getVolumesForTimer(timer *vanusv1alpha1.Timer) []corev1.Volume {
	defaultVolumes := []corev1.Volume{{
		Name: cons.TimerConfigMapName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cons.TimerConfigMapName,
				},
			}},
	}}
	return defaultVolumes
}

func labelsForTimer(name string) map[string]string {
	return map[string]string{"app": name}
}

func annotationsForTimer() map[string]string {
	return map[string]string{"vanus.dev/metrics.port": "2112"}
}
