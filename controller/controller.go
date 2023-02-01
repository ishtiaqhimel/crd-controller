package controller

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/util/intstr"
	"log"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	controllerv1 "github.com/ishtiaqhimel/crd-controller/pkg/apis/crd.com/v1"
	clientset "github.com/ishtiaqhimel/crd-controller/pkg/client/clientset/versioned"
	informer "github.com/ishtiaqhimel/crd-controller/pkg/client/informers/externalversions/crd.com/v1"
	lister "github.com/ishtiaqhimel/crd-controller/pkg/client/listers/crd.com/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	appsinformer "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// Controller is the controller implementation for Ishtiaq resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// sampleclientset is a clientset for our own API group
	sampleclientset clientset.Interface

	deploymentsLister appslisters.DeploymentLister
	deploymentsSynced cache.InformerSynced
	ishtiaqLister     lister.IshtiaqLister
	ishtiaqSynced     cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workQueue workqueue.RateLimitingInterface
}

// NewController returns a new sample controller
func NewController(
	kubeclientset kubernetes.Interface,
	sampleclientset clientset.Interface,
	deploymentInformer appsinformer.DeploymentInformer,
	ishtiaqInformer informer.IshtiaqInformer) *Controller {

	ctrl := &Controller{
		kubeclientset:     kubeclientset,
		sampleclientset:   sampleclientset,
		deploymentsLister: deploymentInformer.Lister(),
		deploymentsSynced: deploymentInformer.Informer().HasSynced,
		ishtiaqLister:     ishtiaqInformer.Lister(),
		ishtiaqSynced:     ishtiaqInformer.Informer().HasSynced,
		workQueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Ishtiaqs"),
	}

	log.Println("Setting up event handlers")

	// Set up an event handler for when Ishtiaq resources change
	ishtiaqInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ctrl.enqueueIshtiaq,
		UpdateFunc: func(oldObj, newObj interface{}) {
			ctrl.enqueueIshtiaq(newObj)
		},
		DeleteFunc: func(obj interface{}) {
			ctrl.enqueueIshtiaq(obj)
		},
	})

	// whatif deployment resources changes??

	return ctrl
}

// enqueueIshtiaq takes an Ishtiaq resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Ishtiaq.
func (c *Controller) enqueueIshtiaq(obj interface{}) {
	log.Println("Enqueueing Ishtiaq...")
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workQueue.AddRateLimited(key)
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shut down the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workQueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	log.Println("Starting Ishtiaq Controller")

	// Wait for the caches to be synced before starting workers
	log.Println("Waiting for informer caches to sync")

	if ok := cache.WaitForCacheSync(stopCh, c.deploymentsSynced, c.ishtiaqSynced); !ok {
		return fmt.Errorf("failed to wait for cache to sync")
	}

	log.Println("Starting workers")

	// Launch two workers to process Ishtiaq resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	log.Println("Worker Started")
	<-stopCh
	log.Println("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the workqueue.
func (c *Controller) runWorker() {
	for c.ProcessNextItem() {
	}
}

func (c *Controller) ProcessNextItem() bool {
	obj, shutdown := c.workQueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func, so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off period.
		defer c.workQueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workQueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}

		// Run the syncHandler, passing it the namespace/name string of the
		// Ishtiaq resource to be synced.
		if err := c.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workQueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}

		// Finally, if no error occurs we Forget this item, so it does not
		// get queued again until another change happens.
		c.workQueue.Forget(obj)
		log.Printf("successfully synced '%s'\n", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}
	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Ishtiaq resource
// with the current status of the resource.
// implement the business logic here.
func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the Ishtiaq resource with this namespace/name
	ishtiaq, err := c.ishtiaqLister.Ishtiaqs(namespace).Get(name)
	if err != nil {
		// The Ishtiaq resource may no longer exist, in which case we stop processing.
		if errors.IsNotFound(err) {
			// We choose to absorb the error here as the worker would requeue the
			// resource otherwise. Instead, the next time the resource is updated
			// the resource will be queued again.
			utilruntime.HandleError(fmt.Errorf("ishtiaq '%s' in work queue no longer exists", key))
			return nil
		}
		return err
	}

	deploymentName := ishtiaq.Spec.Name
	if deploymentName == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		utilruntime.HandleError(fmt.Errorf("%s : deployment name must be specified", key))
		return nil
	}

	// Get the deployment with the name specified in Ishtiaq.spec
	deployment, err := c.deploymentsLister.Deployments(namespace).Get(deploymentName)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		deployment, err = c.kubeclientset.AppsV1().Deployments(ishtiaq.Namespace).Create(context.TODO(), newDeployment(ishtiaq), metav1.CreateOptions{})
	}
	// If an error occurs during Get/Create, we'll requeue the item, so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// If this number of the replicas on the Ishtiaq resource is specified, and the
	// number does not equal the current desired replicas on the Deployment, we
	// should update the Deployment resource.
	if ishtiaq.Spec.Replicas != nil && *ishtiaq.Spec.Replicas != *deployment.Spec.Replicas {
		log.Printf("Ishtiaq %s replicas: %d, deployment replicas: %d\n", name, *ishtiaq.Spec.Replicas, *deployment.Spec.Replicas)

		deployment, err = c.kubeclientset.AppsV1().Deployments(namespace).Update(context.TODO(), newDeployment(ishtiaq), metav1.UpdateOptions{})
		// If an error occurs during Update, we'll requeue the item, so we can
		// attempt processing again later. This could have been caused by a
		// temporary network failure, or any other transient reason.
		if err != nil {
			return err
		}
	}

	// Finally, we update the status block of the Ishtiaq resource to reflect the
	// current state of the world
	err = c.updateIshtiaqStatus(ishtiaq, deployment)
	if err != nil {
		return err
	}

	serviceName := ishtiaq.Spec.Name + "-service"

	// Check if service already exists or not
	service, err := c.kubeclientset.CoreV1().Services(ishtiaq.Namespace).Get(context.TODO(), serviceName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		service, err = c.kubeclientset.CoreV1().Services(ishtiaq.Namespace).Create(context.TODO(), newService(ishtiaq), metav1.CreateOptions{})
		if err != nil {
			log.Println(err)
			return err
		}
		log.Printf("\nservice %s created .....\n", service.Name)
	} else if err != nil {
		log.Println(err)
		return err
	}

	_, err = c.kubeclientset.CoreV1().Services(ishtiaq.Namespace).Update(context.TODO(), service, metav1.UpdateOptions{})
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (c *Controller) updateIshtiaqStatus(ishtiaq *controllerv1.Ishtiaq, deployment *appsv1.Deployment) error {

	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	ishtiaqCopy := ishtiaq.DeepCopy()
	ishtiaqCopy.Status.AvailableReplicas = deployment.Status.AvailableReplicas

	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the Foo resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.
	_, err := c.sampleclientset.CrdV1().Ishtiaqs(ishtiaq.Namespace).Update(context.TODO(), ishtiaqCopy, metav1.UpdateOptions{})

	return err

}

// newDeployment creates a new Deployment for a Ishtiaq resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the Ishtiaq resource that 'owns' it.
func newDeployment(ishtiaq *controllerv1.Ishtiaq) *appsv1.Deployment {

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ishtiaq.Spec.Name,
			Namespace: ishtiaq.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ishtiaq.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "my-app",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "my-app",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "my-app",
							Image: ishtiaq.Spec.Container.Image,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: ishtiaq.Spec.Container.Port,
								},
							},
						},
					},
				},
			},
		},
	}
}

func newService(ishtiaq *controllerv1.Ishtiaq) *corev1.Service {
	labels := map[string]string{
		"app": "my-app",
	}
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind: "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: ishtiaq.Spec.Name + "-service",
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(ishtiaq, controllerv1.SchemeGroupVersion.WithKind("Ishtiaq")),
			},
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeNodePort,
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Port:       ishtiaq.Spec.Container.Port,
					TargetPort: intstr.FromInt(int(ishtiaq.Spec.Container.Port)),
					NodePort:   30007,
				},
			},
		},
	}

}
