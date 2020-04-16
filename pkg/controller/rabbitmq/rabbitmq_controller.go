package rabbitmq

import (
	"context"

	rabbitmqv1alpha1 "github.com/toha10/rabbitmq-operator/pkg/apis/rabbitmq/v1alpha1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_rabbitmq")

// Add creates a new RabbitMQ Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileRabbitMQ{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("rabbitmq-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource RabbitMQ
	err = c.Watch(&source.Kind{Type: &rabbitmqv1alpha1.RabbitMQ{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pods and requeue the owner RabbitMQ
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &rabbitmqv1alpha1.RabbitMQ{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Service and requeue the owner RabbitMQ
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &rabbitmqv1alpha1.RabbitMQ{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Statefulset and requeue the owner RabbitMQ
	err = c.Watch(&source.Kind{Type: &v1.StatefulSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &rabbitmqv1alpha1.RabbitMQ{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileRabbitMQ implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileRabbitMQ{}

// ReconcileRabbitMQ reconciles a RabbitMQ object
type ReconcileRabbitMQ struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a RabbitMQ object and makes changes based on the state read
// and what is in the RabbitMQ.Spec
func (r *ReconcileRabbitMQ) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling RabbitMQ")

	// Fetch the RabbitMQ instance
	instance := &rabbitmqv1alpha1.RabbitMQ{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
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
	if err := controllerutil.SetControllerReference(instance, cm, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this ConfigMap already exists
	foundCM := &corev1.ConfigMap{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}, foundCM)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new ConfigMap", "ConfigMap.Namespace", cm.Namespace, "Pod.Name", cm.Name)
		err = r.client.Create(context.TODO(), cm)
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
	if err := controllerutil.SetControllerReference(instance, rmqService, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Service already exists
	foundRMQService := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: rmqService.Name, Namespace: rmqService.Namespace}, foundRMQService)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Service", "StatefulSet.Namespace", rmqService.Namespace, "Pod.Name", rmqService.Name)
		err = r.client.Create(context.TODO(), rmqService)
		if err != nil {
			return reconcile.Result{}, err
		}
		// Service created successfully - don't requeue
	} else if err != nil {
		return reconcile.Result{}, err
	} else {
		reqLogger.Info("Skip reconcile: Service already exists", "Service.Namespace", foundRMQService.Namespace, "Service.Name", foundRMQService.Name)
	}

	// Define a new StatefulSet object
	ss := newStatefulSet(instance)

	// Set RabbitMQ instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, ss, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this StatefulSet already exists
	foundSS := &v1.StatefulSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: ss.Name, Namespace: ss.Namespace}, foundSS)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new StatefulSet", "StatefulSet.Namespace", ss.Namespace, "Pod.Name", ss.Name)
		err = r.client.Create(context.TODO(), ss)
		if err != nil {
			return reconcile.Result{}, err
		}
		// StatefulSet created successfully - don't requeue
	} else if err != nil {
		return reconcile.Result{}, err
	} else {
		reqLogger.Info("Skip reconcile: StatefulSet already exists", "StatefulSet.Namespace", foundSS.Namespace, "StatefulSet.Name", foundSS.Name)
	}

	return reconcile.Result{}, nil
}
