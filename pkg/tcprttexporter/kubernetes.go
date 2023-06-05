package tcprttexporter

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	podIPIndex = "PodIP"
)

// IPResolver is used to resolve ip address to human readable representation(e.g. kubernetes service/pod namespaced name)
type IPResolver interface {
	Resolve(ip string) string
}

var _ IPResolver = &kubernetesIPResolver{}

type kubernetesIPResolver struct {
	podLister     cache.Indexer
	serviceLister cache.Indexer
}

func NewKubernetesIPResolver(kubeconfig string) (IPResolver, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	kubeclient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	podLister, podController := cache.NewIndexerInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return kubeclient.CoreV1().Pods(v1.NamespaceAll).List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return kubeclient.CoreV1().Pods(v1.NamespaceAll).Watch(context.TODO(), options)
			},
		},
		&corev1.Pod{},
		0,
		cache.ResourceEventHandlerFuncs{
			// don't care, nothing need to be updated
			AddFunc:    func(obj interface{}) {},
			UpdateFunc: func(oldObj, newObj interface{}) {},
			DeleteFunc: func(obj interface{}) {},
		},
		cache.Indexers{
			podIPIndex: func(obj interface{}) ([]string, error) {
				pod := obj.(*corev1.Pod)
				return []string{pod.Status.PodIP}, nil
			},
		})
	go podController.Run(context.Background().Done())

	serviceLister, serviceController := cache.NewIndexerInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return kubeclient.CoreV1().Services(v1.NamespaceAll).List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return kubeclient.CoreV1().Services(v1.NamespaceAll).Watch(context.TODO(), options)
			},
		},
		&corev1.Service{},
		0,
		cache.ResourceEventHandlerFuncs{
			// don't care, nothing need to be updated
			AddFunc:    func(obj interface{}) {},
			UpdateFunc: func(oldObj, newObj interface{}) {},
			DeleteFunc: func(obj interface{}) {},
		},
		cache.Indexers{
			podIPIndex: func(obj interface{}) ([]string, error) {
				service := obj.(*corev1.Service)
				return []string{service.Spec.ClusterIP}, nil
			},
		})
	go serviceController.Run(context.Background().Done())

	return &kubernetesIPResolver{
		podLister:     podLister,
		serviceLister: serviceLister,
	}, nil
}

func (r *kubernetesIPResolver) Resolve(ip string) string {
	objs, err := r.podLister.ByIndex(podIPIndex, ip)
	if err == nil && len(objs) > 0 {
		pod := objs[0].(*corev1.Pod)
		return fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
	}
	objs, err = r.serviceLister.ByIndex(podIPIndex, ip)
	if err == nil && len(objs) > 0 {
		service := objs[0].(*corev1.Service)
		return fmt.Sprintf("%s/%s", service.Namespace, service.Name)
	}
	return ip
}
