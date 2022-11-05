package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	Client   *kubernetes.Clientset
	Informer cache.SharedIndexInformer
	Queue    workqueue.RateLimitingInterface
}

func NewController(client *kubernetes.Clientset,
	informerFactory informers.SharedInformerFactory) *Controller {
	deploymentInformer := informerFactory.Apps().V1().Deployments().Informer()
	c := &Controller{
		Client:   client,
		Informer: deploymentInformer,
		Queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(),
			"myqueue"),
	}
	deploymentInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handleAdd,
		DeleteFunc: c.handleDelete,
	})
	return c
}

func (c *Controller) handleAdd(obj interface{}) {
	fmt.Println("Handling deployment add")
	c.Queue.Add(obj)
}

func (c *Controller) handleDelete(obj interface{}) {
	fmt.Println("Handling deployment delete")
	c.Queue.Add(obj)
}

func (c *Controller) Run(ch <-chan struct{}) {
	fmt.Println("starting controller")
	if !cache.WaitForCacheSync(ch, c.Informer.HasSynced) {
		fmt.Print("waiting for cache to be synced\n")
	}
	go wait.Until(c.worker, 1*time.Second, ch)
	<-ch
}

func (c *Controller) worker() {
	for c.processItems() {
	}
}

func (c *Controller) processItems() bool {
	item, shutdown := c.Queue.Get()
	if shutdown {
		return false
	}
	c.Queue.Forget(item)
	bgContext := context.Background()
	deployment := item.(*appsv1.Deployment)

	_, err := c.Client.AppsV1().Deployments(deployment.Namespace).Get(
		bgContext, deployment.Name, metav1.GetOptions{})
	if err != nil && strings.Contains(err.Error(), "not found") {
		serviceError := c.deleteService(bgContext, deployment)
		if serviceError != nil {
			fmt.Println(serviceError.Error())
			return false
		}
		fmt.Printf("Deleted service for deployment %s\n", deployment.Name)
	} else {
		serviceError := c.createService(bgContext, deployment)
		if serviceError != nil {
			fmt.Println(serviceError.Error())
			return false
		}
		fmt.Printf("Created service for deployment %s\n", deployment.Name)
	}
	return true
}

func (c *Controller) createService(bgContext context.Context,
	deployment *appsv1.Deployment) error {
	_, err := c.Client.CoreV1().Services(deployment.Namespace).Get(
		bgContext, ServiceName(deployment), metav1.GetOptions{})
	if err == nil {
		// service for deployment found, don't create it again
		return err
	}
	fmt.Printf("creating service for deployment %s\n", deployment.Name)
	servicePorts := ServicePorts(deployment)
	podLabels, err := PodLabels(bgContext, c.Client, deployment)
	if err != nil {
		return err
	}
	_, err = c.Client.CoreV1().Services(deployment.Namespace).Create(
		bgContext,
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ServiceName(deployment),
				Namespace: deployment.Namespace,
				Labels:    podLabels,
			},
			Spec: corev1.ServiceSpec{
				Ports:    servicePorts,
				Selector: podLabels,
				Type:     "LoadBalancer",
			},
		}, metav1.CreateOptions{})
	return err
}

func (c *Controller) deleteService(bgContext context.Context, deployment *appsv1.Deployment) error {
	serviceName := ServiceName(deployment)
	fmt.Printf("deleting service %s\n", serviceName)
	return c.Client.CoreV1().Services(deployment.Namespace).Delete(
		bgContext, serviceName, metav1.DeleteOptions{})
}
