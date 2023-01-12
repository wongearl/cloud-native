package k8sunit

import (
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
)

// Add adds or updates the given Event object.
func Add(kubeClient kubernetes.Interface, eventObj *corev1.Event) error {

	return nil
}
func GetEvent() *corev1.Event {
	return nil
}

// Create creates the given CRD objects or updates them if these objects already exist in the cluster.
func Create(client clientset.Interface, crds ...*apiextensionsv1.CustomResourceDefinition) error {
	return nil
}

func GetStorageClass() *storagev1.StorageClass {
	return nil
}
