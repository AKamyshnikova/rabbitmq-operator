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
	"reflect"

	rabbitmqv1alpha1 "github.com/toha10/rabbitmq-operator/api/v1alpha1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("controller_rabbitmq")

// RabbitMQReconciler reconciles a RabbitMQ object
type RabbitMQReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=rabbitmq.mirantis.com,resources=rabbitmqs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rabbitmq.mirantis.com,resources=rabbitmqs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=rabbitmq.mirantis.com,resources=rabbitmqs/finalizers,verbs=update
func (r *RabbitMQReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	reqLogger.Info("Reconciling RabbitMQ")

	// Fetch the RabbitMQ instance
	instance := &rabbitmqv1alpha1.RabbitMQ{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Define a new ConfigMap object
	cm := newConfigMap(instance)

	// Set RabbitMQ instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, cm, r.Scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this ConfigMap already exists
	foundCM := &corev1.ConfigMap{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}, foundCM)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new ConfigMap", "ConfigMap.Namespace", cm.Namespace, "Pod.Name", cm.Name)
		err = r.Client.Create(context.TODO(), cm)
		if err != nil {
			reqLogger.Error(err, "Config Map creation has been failed")
			return reconcile.Result{}, err
		}
		// ConfigMap created successfully - don't requeue
	} else if err != nil {
		return reconcile.Result{}, err
	} else {
		reqLogger.Info("Skip reconcile: ConfigMap already exists", "ConfigMap.Namespace", foundCM.Namespace, "StatefulSet.Name", foundCM.Name)
	}

	// Define a new Service object
	rmqService := newService(instance)

	// Set RabbitMQ instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, rmqService, r.Scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Service already exists
	foundRMQService := &corev1.Service{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: rmqService.Name, Namespace: rmqService.Namespace}, foundRMQService)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Service", "StatefulSet.Namespace", rmqService.Namespace, "Pod.Name", rmqService.Name)
		err = r.Client.Create(context.TODO(), rmqService)
		if err != nil {
			return reconcile.Result{}, err
		}
		// Service created successfully - don't requeue
	} else if err != nil {
		return reconcile.Result{}, err
	} else {
		reqLogger.Info("Skip reconcile: Service already exists", "Service.Namespace", foundRMQService.Namespace, "Service.Name", foundRMQService.Name)
	}

	// Define a new Headless Service object
	rmqHeadlessService := newHeadlessService(instance)

	// Set RabbitMQ instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, rmqHeadlessService, r.Scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Service already exists
	foundRMQHeadlessService := &corev1.Service{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: rmqHeadlessService.Name, Namespace: rmqHeadlessService.Namespace}, foundRMQHeadlessService)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Service", "Service.Namespace", rmqHeadlessService.Namespace, "Service.Name", rmqHeadlessService.Name)
		err = r.Client.Create(context.TODO(), rmqHeadlessService)
		if err != nil {
			return reconcile.Result{}, err
		}
		// Service created successfully - don't requeue
	} else if err != nil {
		return reconcile.Result{}, err
	} else {
		reqLogger.Info("Skip reconcile: Service already exists", "Service.Namespace", foundRMQHeadlessService.Namespace, "Service.Name", foundRMQHeadlessService.Name)
	}

	if instance.Spec.ExporterPort > 0 {
		// Define a new Exporter Service object
		rmqExporterService := newExporterService(instance)

		// Set RabbitMQ instance as the owner and controller
		if err := controllerutil.SetControllerReference(instance, rmqExporterService, r.Scheme); err != nil {
			return reconcile.Result{}, err
		}

		// Check if this Exporter Service already exists
		foundrmqExporterService := &corev1.Service{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: rmqExporterService.Name, Namespace: rmqExporterService.Namespace}, foundrmqExporterService)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new Service", "Service.Namespace", rmqExporterService.Namespace, "Service.Name", rmqExporterService.Name)
			err = r.Client.Create(context.TODO(), rmqExporterService)
			if err != nil {
				return reconcile.Result{}, err
			}
			// Service created successfully - don't requeue
		} else if err != nil {
			return reconcile.Result{}, err
		} else {
			reqLogger.Info("Skip reconcile: Service already exists", "Service.Namespace", foundrmqExporterService.Namespace, "Service.Name", foundrmqExporterService.Name)
		}
	}

	// Define a new StatefulSet object
	ss := newStatefulSet(instance)

	// Set RabbitMQ instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, ss, r.Scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this StatefulSet already exists
	foundSS := &v1.StatefulSet{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: ss.Name, Namespace: ss.Namespace}, foundSS)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new StatefulSet", "StatefulSet.Namespace", ss.Namespace, "Pod.Name", ss.Name)
		err = r.Client.Create(context.TODO(), ss)
		if err != nil {
			return reconcile.Result{}, err
		}
		// StatefulSet created successfully - don't requeue
	} else if err != nil {
		return reconcile.Result{}, err
	}

	if reflect.DeepEqual(foundSS.Spec, ss.Spec) {
		reqLogger.Info("RabbitMQ StatefulSet already exists and looks updated", "Name", foundSS.Name)
	} else {
		reqLogger.Info("Update RabbitMQ StatefulSet", "Namespace", ss.Namespace, "Name", ss.Name)
		ss.ObjectMeta.ResourceVersion = foundSS.ObjectMeta.ResourceVersion
		err = r.Client.Update(context.TODO(), ss)
		if err != nil {
			reqLogger.Error(err, "RabbitMQ StatefulSet cannot be updated")
			return reconcile.Result{}, err
		}
	}

	if instance.Spec.ExporterPort > 0 {
		exporter := newDeployment(instance)

		if err := controllerutil.SetControllerReference(instance, exporter, r.Scheme); err != nil {
			return reconcile.Result{}, err
		}

		// Check if this StatefulSet already exists
		foundExporter := &v1.Deployment{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: exporter.Name, Namespace: exporter.Namespace}, foundExporter)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new Exporter Deployment", "Deployment.Namespace", exporter.Namespace, "Deployment.Name", exporter.Name)
			err = r.Client.Create(context.TODO(), exporter)
			if err != nil {
				return reconcile.Result{}, err
			}
			// Deployment created successfully - don't requeue
		} else if err != nil {
			return reconcile.Result{}, err
		}

		if reflect.DeepEqual(foundExporter.Spec, exporter.Spec) {
			reqLogger.Info("RabbitMQ Exporter Deployment already exists and looks updated", "Name", foundExporter.Name)
		} else {
			reqLogger.Info("Update RabbitMQ Exporter Deployment", "Namespace", exporter.Namespace, "Name", exporter.Name)
			exporter.ObjectMeta.ResourceVersion = foundExporter.ObjectMeta.ResourceVersion
			err = r.Client.Update(context.TODO(), exporter)
			if err != nil {
				reqLogger.Error(err, "RabbitMQ Exporter Deployment cannot be updated")
				return reconcile.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RabbitMQReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&rabbitmqv1alpha1.RabbitMQ{}).
		Owns(&corev1.Pod{}).
		Owns(&corev1.Service{}).
		Owns(&v1.StatefulSet{}).
		Complete(r)
}
