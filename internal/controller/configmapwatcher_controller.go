/*
Copyright 2025.

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

package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1alpha1 "github.com/HunnTeRUS/meu-operator/api/v1alpha1"
)

// ConfigMapWatcherReconciler reconciles a ConfigMapWatcher object
type ConfigMapWatcherReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps.exemplo.com,resources=configmapwatchers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.exemplo.com,resources=configmapwatchers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.exemplo.com,resources=configmapwatchers/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch

// Reconcile handles the reconciliation loop for ConfigMapWatcher resources
func (r *ConfigMapWatcherReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the ConfigMapWatcher instance
	configMapWatcher := &appsv1alpha1.ConfigMapWatcher{}
	err := r.Get(ctx, req.NamespacedName, configMapWatcher)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("ConfigMapWatcher resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get ConfigMapWatcher")
		return ctrl.Result{}, err
	}

	// Fetch the target ConfigMap
	configMap := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{
		Name:      configMapWatcher.Spec.ConfigMapName,
		Namespace: configMapWatcher.Spec.ConfigMapNamespace,
	}, configMap)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Target ConfigMap not found", "ConfigMap", configMapWatcher.Spec.ConfigMapName)
			return ctrl.Result{RequeueAfter: time.Minute}, nil
		}
		log.Error(err, "Failed to get target ConfigMap")
		return ctrl.Result{}, err
	}

	// Check if ConfigMap has changed
	currentVersion := configMap.ResourceVersion
	if currentVersion == configMapWatcher.Status.LastConfigMapVersion {
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	// Prepare event data
	eventData := map[string]interface{}{
		"configMapName":      configMap.Name,
		"configMapNamespace": configMap.Namespace,
		"resourceVersion":    configMap.ResourceVersion,
		"data":               configMap.Data,
		"binaryData":         configMap.BinaryData,
		"timestamp":          time.Now().UTC().Format(time.RFC3339),
	}

	// Get credentials if specified
	var httpClient *http.Client
	if configMapWatcher.Spec.EventSecretName != "" {
		secret := &corev1.Secret{}
		err = r.Get(ctx, types.NamespacedName{
			Name:      configMapWatcher.Spec.EventSecretName,
			Namespace: configMapWatcher.Spec.EventSecretNamespace,
		}, secret)
		if err != nil {
			log.Error(err, "Failed to get event secret")
			return ctrl.Result{}, err
		}
		// TODO: Implement proper authentication using the secret
		httpClient = &http.Client{}
	} else {
		httpClient = &http.Client{}
	}

	// Send event
	jsonData, err := json.Marshal(eventData)
	if err != nil {
		log.Error(err, "Failed to marshal event data")
		return ctrl.Result{}, err
	}

	resp, err := httpClient.Post(configMapWatcher.Spec.EventEndpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Error(err, "Failed to send event")
		return ctrl.Result{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to send event: status code %d", resp.StatusCode)
		log.Error(err, "Event sending failed")
		return ctrl.Result{}, err
	}

	// Update status
	configMapWatcher.Status.LastConfigMapVersion = currentVersion
	configMapWatcher.Status.LastEventSent = metav1.Now()
	if err := r.Status().Update(ctx, configMapWatcher); err != nil {
		log.Error(err, "Failed to update ConfigMapWatcher status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: time.Minute}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConfigMapWatcherReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.ConfigMapWatcher{}).
		Watches(
			&corev1.ConfigMap{},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForConfigMap),
		).
		Complete(r)
}

// findObjectsForConfigMap finds ConfigMapWatcher objects that are watching the given ConfigMap
func (r *ConfigMapWatcherReconciler) findObjectsForConfigMap(ctx context.Context, obj client.Object) []reconcile.Request {
	configMap := obj.(*corev1.ConfigMap)
	var requests []reconcile.Request

	var watchers appsv1alpha1.ConfigMapWatcherList
	if err := r.List(ctx, &watchers); err != nil {
		return requests
	}

	for _, watcher := range watchers.Items {
		if watcher.Spec.ConfigMapName == configMap.Name && watcher.Spec.ConfigMapNamespace == configMap.Namespace {
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      watcher.Name,
					Namespace: watcher.Namespace,
				},
			})
		}
	}

	return requests
}
