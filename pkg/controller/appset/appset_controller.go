package appset

import (
	"context"
	"reflect"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	appv1alpha1 "github.com/kube_op/minicd-operator/pkg/apis/app/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_appset")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new AppSet Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileAppSet{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("appset-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource AppSet
	err = c.Watch(&source.Kind{Type: &appv1alpha1.AppSet{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner AppSet
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1alpha1.AppSet{},
	})
	if err != nil {
		return err
	}
	
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appv1alpha1.AppSet{},
	})
	if err != nil {
		return err
	}	

	return nil
}

// blank assignment to verify that ReconcileAppSet implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileAppSet{}

// ReconcileAppSet reconciles a AppSet object
type ReconcileAppSet struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a AppSet object and makes changes based on the state read
// and what is in the AppSet.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileAppSet) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling AppSet")

	// Fetch the AppSet instance
	instance := &appv1alpha1.AppSet{}
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
	// List all pods owned by this PodSet instance
	lbls := labels.Set{
		"app":     instance.Name,
	}
	existingPods := &corev1.PodList{}
	err = r.client.List(context.TODO(),
		existingPods,
		&client.ListOptions{
			Namespace:     request.Namespace,
			LabelSelector: labels.SelectorFromSet(lbls),
		})
	if err != nil {
		reqLogger.Error(err, "failed to list existing pods in the instance")
		return reconcile.Result{}, err
	}
	existingPodNames := []string{}
	// Count the pods that are pending or running as available
	for _, pod := range existingPods.Items {
		if pod.GetObjectMeta().GetDeletionTimestamp() != nil {
			continue
		}
		if pod.Status.Phase == corev1.PodPending || pod.Status.Phase == corev1.PodRunning {
			existingPodNames = append(existingPodNames, pod.GetObjectMeta().GetName())
		}
	}

	reqLogger.Info("Checking instance", "expected replicas", instance.Spec.Replicas, "Pod.Names", existingPodNames)
	// Update the status if necessary
	status := appv1alpha1.AppSetStatus{
		Replicas: int32(len(existingPodNames)),
		PodNames: existingPodNames,
	}
	if !reflect.DeepEqual(instance.Status, status) {
		instance.Status = status
		err := r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "failed to update the instance")
			return reconcile.Result{}, err
		}
	}

	// Define a new Pod object
	//pod := newPodForCR(instance)
	deployment := newDeploymentForCR(instance)
	service := newClusterIPServiceForCR(instance)
			//Type:            cr.Spec.ServiceType,
	if strings.EqualFold(instance.Spec.ServiceType, "clusterip"){
     service = newClusterIPServiceForCR(instance)
	}
	if strings.EqualFold(instance.Spec.ServiceType, "nodeport"){
		service = newNodePortServiceForCR(instance)
	}
	if strings.EqualFold(instance.Spec.ServiceType, "loadbalancer"){
		service = newLoadBalancerServiceForCR(instance)
	}
	
	//service := newServiceForCR(instance)

	// Set AppSet instance as the owner and controller

	if err := controllerutil.SetControllerReference(instance, deployment, r.scheme); err != nil {
		return reconcile.Result{}, err
	}
	if err := controllerutil.SetControllerReference(instance, service, r.scheme); err != nil {
		return reconcile.Result{}, err
	}	

	// Check if this Service already exists
	founds := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, founds)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
		err = r.client.Create(context.TODO(), service)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Service created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}


		// Check if this Deployment already exists
	foundd := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, foundd)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		err = r.client.Create(context.TODO(), deployment)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Deployment created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}
		
	reqLogger.Info("Updating Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
	err = r.client.Update(context.TODO(), deployment)
	if err != nil {
		return reconcile.Result{}, err
	}
	//reqLogger.Info("Updating Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
	//err = r.client.Update(context.TODO(), service)
	//if err != nil {
		//return reconcile.Result{}, err
	//}

	// Success - don't requeue
	return reconcile.Result{}, nil

}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *appv1alpha1.AppSet) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}

// newDeploymentForCR returns a busybox pod with the same name/namespace as the cr
func newDeploymentForCR(cr *appv1alpha1.AppSet) *appsv1.Deployment {
	return 	&appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: cr.Name + "-deployment",
			Namespace: cr.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(cr.Spec.Replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": cr.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": cr.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  cr.Name + "-pod",
							Image: cr.Spec.Image,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 8080,
								},
							},
						},
					},
				},
			},
		},
	}

}

func newLoadBalancerServiceForCR(cr *appv1alpha1.AppSet) *corev1.Service {

	selectorLabels := generateSelectorLabels(cr.Name)

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
		Name: cr.Name + "-service",
		Namespace: cr.Namespace,
	},
	Spec: corev1.ServiceSpec{
		Ports: []corev1.ServicePort{
			{Name: cr.Name + "-service", Port: 8080, Protocol: "TCP", TargetPort: intstr.FromInt(8080)},
		},
		Selector: selectorLabels,
		SessionAffinity: corev1.ServiceAffinityNone,
		Type:            corev1.ServiceTypeLoadBalancer,
	},
}
}

func newNodePortServiceForCR(cr *appv1alpha1.AppSet) *corev1.Service {

	selectorLabels := generateSelectorLabels(cr.Name)

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
		Name: cr.Name + "-service",
		Namespace: cr.Namespace,
	},
	Spec: corev1.ServiceSpec{
		Ports: []corev1.ServicePort{
			{Name: cr.Name + "-service", Port: 8080, Protocol: "TCP", TargetPort: intstr.FromInt(8080)},
		},
		Selector: selectorLabels,
		SessionAffinity: corev1.ServiceAffinityNone,
		Type:            corev1.ServiceTypeNodePort,
	},
}
}

func newClusterIPServiceForCR(cr *appv1alpha1.AppSet) *corev1.Service {

	selectorLabels := generateSelectorLabels(cr.Name)

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
		Name: cr.Name + "-service",
		Namespace: cr.Namespace,
	},
	Spec: corev1.ServiceSpec{
		Ports: []corev1.ServicePort{
			{Name: cr.Name + "-service", Port: 8080, Protocol: "TCP", TargetPort: intstr.FromInt(8080)},
		},
		Selector: selectorLabels,
		SessionAffinity: corev1.ServiceAffinityNone,
		Type:            corev1.ServiceTypeClusterIP,
	},
}
}

func int32Ptr(i int32) *int32 { return &i }

func generateSelectorLabels( name string) map[string]string {
	return map[string]string{
		"app":      name,
	}
}