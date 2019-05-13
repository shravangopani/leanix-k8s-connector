package main

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// KubernetesAPI is an optionated facade for the kubernetes api
type KubernetesAPI struct {
	Client kubernetes.Interface
}

// NewKubernetesAPI creates a new kuberntes api client
func NewKubernetesAPI(config *rest.Config) (*KubernetesAPI, error) {
	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &KubernetesAPI{
		Client: clientset,
	}, nil
}

// Deployments returns a list of deployments filted by the given blacklisted namespaces
func (k *KubernetesAPI) Deployments(blacklistedNamespaces []string) (*appsv1.DeploymentList, error) {
	fieldSelector := BlacklistFieldSelector(blacklistedNamespaces)
	deployments, err := k.Client.AppsV1().Deployments("").List(metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, err
	}
	return deployments, nil
}

// DeploymentsOnNodes returns a list of deployments filted by the given blacklisted namespaces
// and the nodes the deploment pods are running on
func (k *KubernetesAPI) DeploymentsOnNodes(blacklistedNamespaces []string) (*appsv1.DeploymentList, *[]corev1.Node, error) {
	deployments, err := k.Deployments(blacklistedNamespaces)
	if err != nil {
		return nil, nil, err
	}
	nodeNames := NewStringSet()
	for _, d := range deployments.Items {
		pods, err := k.Pods(&d)
		if err != nil {
			return nil, nil, err
		}
		for _, p := range pods.Items {
			nodeNames.Add(p.Spec.NodeName)
		}
	}
	nodes, err := k.Nodes()
	if err != nil {
		return nil, nil, err
	}
	deploymentNodes := make([]corev1.Node, 0)
	for _, n := range nodes.Items {
		if nodeNames.Contains(n.Name) {
			deploymentNodes = append(deploymentNodes, n)
		}
	}
	return deployments, &deploymentNodes, nil
}

// Pods retuns a list of pods matching the selectors of the given deployment
func (k *KubernetesAPI) Pods(deployment *appsv1.Deployment) (*corev1.PodList, error) {
	labelMatcher := make([]string, 0)
	for label, val := range deployment.Spec.Selector.MatchLabels {
		labelMatcher = append(labelMatcher, fmt.Sprintf("%s=%s", label, val))
	}
	selector := strings.Join(labelMatcher, ",")
	pods, err := k.Client.CoreV1().Pods("").List(metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, err
	}
	return pods, nil
}

// StatefulSets returns a list of statefulsets filted by the given blacklisted namespaces
func (k *KubernetesAPI) StatefulSets(blacklistedNamespaces []string) (*appsv1.StatefulSetList, error) {
	fieldSelector := BlacklistFieldSelector(blacklistedNamespaces)
	statefulsets, err := k.Client.AppsV1().StatefulSets("").List(metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, err
	}
	return statefulsets, nil
}

// Nodes gets the list of worker nodes (kubelets)
func (k *KubernetesAPI) Nodes() (*corev1.NodeList, error) {
	nodes, err := k.Client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

// BlacklistFieldSelector builds a Field Selector string to filter the reponse to not
// include resources, that live in the blacklisted namespaces.
func BlacklistFieldSelector(blacklistedNamespaces []string) string {
	namespaceSelectors := Prefix(blacklistedNamespaces, "metadata.namespace!=")
	return strings.Join(namespaceSelectors, ",")
}

// Prefix return a new list where all items are prefixed with the string given as prefix
func Prefix(l []string, p string) []string {
	r := make([]string, 0)
	for _, e := range l {
		r = append(r, (p + e))
	}
	return r
}
