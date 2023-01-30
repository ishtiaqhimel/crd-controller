package controller

import (
	"log"

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

// Controller is the controller implementation for Foo resources
type Controller struct {
	kubeclientset   kubernetes.Interface
	sampleclientset clientset.Interface

	deploymentsLister appslisters.DeploymentLister
	deploymentsSynced cache.InformerSynced
	ishtiaqLister     lister.IshtiaqLister
	ishtiaqSynced     cache.InformerSynced

	workQueue workqueue.RateLimitingInterface
}

// NewController a constructor
// returns a new
func NewController(
	kubeclientset kubernetes.Interface,
	sampleclientset clientset.Interface,
	deploymentInformer appsinformer.DeploymentInformer,
	ishtiaqInformer informer.IshtiaqInformer) *Controller {

	ctrl := &Controller{
		kubeclientset:     kubeclientset,
		sampleclientset:   sampleclientset,
		deploymentsLister: deploymentInformer.Lister(),
		ishtiaqLister:     ishtiaqInformer.Lister(),
		ishtiaqSynced:     ishtiaqInformer.Informer().HasSynced,
		workQueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Ishtiaqs"),
	}

	log.Println("Setting up event handlers")

	ishtiaqInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: Controller.enqueueIshtiaq,
		UpdateFunc: func(oldObj, newObj interface{}) {
			Controller.enqueueIshtiaq(newObj)
		},
		DeleteFunc: func(obj interface{}) {
			Controller.enqueueIshtiaq(obj)
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
