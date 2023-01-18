package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

func ServiceName(deployment *appsv1.Deployment) string {
	return deployment.Name + "-service"
}

func ServicePorts(deployment *appsv1.Deployment) []corev1.ServicePort {
	servicePorts := []corev1.ServicePort{}
	for _, container := range deployment.Spec.Template.Spec.Containers {
		for _, containerPort := range container.Ports {
			port := corev1.ServicePort{
				Name:     containerPort.Name,
				Protocol: containerPort.Protocol,
				Port:     containerPort.ContainerPort,
			}
			servicePorts = append(servicePorts, port)
		}
	}
	if len(servicePorts) == 0 {
		port := corev1.ServicePort{
			Name:     "default-port",
			Protocol: "tcp",
			Port:     80,
		}
		servicePorts = append(servicePorts, port)
	}
	return servicePorts
}

func PodLabels(bgContext context.Context, client *kubernetes.Clientset,
	deployment *appsv1.Deployment) (map[string]string, error) {
	labelSelector := labels.Set(deployment.Spec.Selector.MatchLabels).AsSelector().String()
	pods, err := client.CoreV1().Pods(deployment.Namespace).List(bgContext,
		metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, err
	}
	for _, pod := range pods.Items {
		return pod.Labels, nil
	}
	return nil, nil
}
