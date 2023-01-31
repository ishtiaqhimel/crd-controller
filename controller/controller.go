package controller

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"log"
	"time"

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
	kubeclientset   kubernetes.Interface
	sampleclientset clientset.Interface

	deploymentsLister appslisters.DeploymentLister
	deploymentsSynced cache.InformerSynced
	ishtiaqLister     lister.IshtiaqLister
	ishtiaqSynced     cache.InformerSynced

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

	ishtiaqInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ctrl.enqueueIshtiaq,
		UpdateFunc: func(oldObj, newObj interface{}) {
			ctrl.enqueueIshtiaq(newObj)
		},
		DeleteFunc: func(obj interface{}) {
			ctrl.enqueueIshtiaq(obj)
		},
	})

	return ctrl
}

func (c *Controller) enqueueIshtiaq(obj interface{}) {
	log.Println("Enqueueing Ishtiaq...")
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workQueue.AddRateLimited(key)
}

func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workQueue.ShutDown()

	log.Println("Starting Ishtiaq Controller")

	log.Println("Waiting for informer caches to sync")

	if c.deploymentsSynced != nil {
		fmt.Println("deploy sycn")
	}
	if c.ishtiaqSynced != nil {
		fmt.Println("istiaq sync")
	}
	if ok := cache.WaitForCacheSync(stopCh, c.deploymentsSynced, c.ishtiaqSynced); !ok {
		return fmt.Errorf("failed to wait for cache to sync")
	}

	log.Println("Starting workers")

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	log.Println("Worker Started")
	<-stopCh
	log.Println("Shutting down workers")

	return nil
}

func (c *Controller) runWorker() {
	for c.ProcessNextItem() {
	}
}

func (c *Controller) ProcessNextItem() bool {
	obj, shutdown := c.workQueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.workQueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			c.workQueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}

		if err := c.syncHandler(key); err != nil {
			c.workQueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}

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

func (c *Controller) syncHandler(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)

	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
	}

	ishtiaq, err := c.ishtiaqLister.Ishtiaqs(namespace).Get(name)

	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("foo '%s' in work queue no longer exists", key))
			return nil
		}
		return err
	}
	deploymentName := ishtiaq.Spec.Name
	if deploymentName == "" {
		utilruntime.HandleError(fmt.Errorf("%s : deployment name must be specified", key))
		return nil
	}

	deployment, err := c.deploymentsLister.Deployments(namespace).Get(deploymentName)

	if errors.IsNotFound(err) {
		deployment, err = c.kubeclientset.AppsV1().Deployments(ishtiaq.Namespace).Create(context.TODO(), newDeployment(ishtiaq), metav1.CreateOptions{})
	}
	if err != nil {
		return err
	}

	if ishtiaq.Spec.Replicas != nil && *ishtiaq.Spec.Replicas != *deployment.Spec.Replicas {
		log.Printf("Foo %s replicas: %d, deployment replicas: %d\n", name, *ishtiaq.Spec.Replicas, *deployment.Spec.Replicas)

		deployment, err = c.kubeclientset.AppsV1().Deployments(namespace).Update(context.TODO(), newDeployment(ishtiaq), metav1.UpdateOptions{})

		if err != nil {
			return err
		}
	}

	err = c.updateIshtiaqStatus(ishtiaq, deployment)
	if err != nil {
		return err
	}

	return nil
}

func (c *Controller) updateIshtiaqStatus(ishtiaq *controllerv1.Ishtiaq, deployment *appsv1.Deployment) error {

	ishtiaqCopy := ishtiaq.DeepCopy()
	ishtiaqCopy.Status.AvailableReplicas = deployment.Status.AvailableReplicas

	_, err := c.sampleclientset.CrdV1().Ishtiaqs(ishtiaq.Namespace).Update(context.TODO(), ishtiaqCopy, metav1.UpdateOptions{})

	return err

}

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
							Image: "ishtiaq99/go-api-server",
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 3000,
								},
							},
						},
					},
				},
			},
		},
	}
}
