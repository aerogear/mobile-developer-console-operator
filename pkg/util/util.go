package util

import (
	"fmt"
	"k8s.io/client-go/discovery"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Labels(cr *metav1.ObjectMeta, suffix string) map[string]string {
	return map[string]string{
		"app":     cr.Name,
		"service": fmt.Sprintf("%s-%s", cr.Name, suffix),
	}
}

// ObjectMeta returns the default ObjectMeta for all the other objects here
func ObjectMeta(cr *metav1.ObjectMeta, suffix string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      fmt.Sprintf("%s-%s", cr.Name, suffix),
		Namespace: cr.Namespace,
		Labels:    Labels(cr, suffix),
	}
}

func GeneratePassword() (string, error) {
	generatedPassword, err := uuid.NewRandom()
	if err != nil {
		return "", errors.Wrap(err, "error generating password")
	}
	return strings.Replace(generatedPassword.String(), "-", "", -1), nil
}

// apiVersionExists checks if a given API version exists in Kubernetes cluster.
// Modified from https://github.com/operator-framework/operator-sdk/blob/947a464dbe968b8af147049e76e40f787ccb0847/pkg/k8sutil/k8sutil.go#L93
// The Operator Framework one checks a specific resource exists, but this function checks if an API version exists.
// Theoretically, there can be 2 resources in an API version, 1 exists and 1 not.
func ApiVersionExists(dc discovery.DiscoveryInterface, apiGroupVersion string) (bool, error) {
	apiLists, err := dc.ServerResources()
	if err != nil {
		return false, err
	}
	for _, apiList := range apiLists {
		if apiList.GroupVersion == apiGroupVersion {
			return true, nil
		}
	}
	return false, nil
}

func FindContainerSpec(deployment *appsv1.Deployment, name string) *corev1.Container {
	if deployment == nil || &deployment.Spec == nil || &deployment.Spec.Template == nil || &deployment.Spec.Template.Spec == nil || &deployment.Spec.Template.Spec.Containers == nil || len(deployment.Spec.Template.Spec.Containers) == 0 {
		return nil
	}

	for _, spec := range deployment.Spec.Template.Spec.Containers {
		if spec.Name == name {
			return &spec
		}
	}

	return nil
}

func UpdateContainerSpecImage(deployment *appsv1.Deployment, name string, image string) {
	if deployment == nil || &deployment.Spec == nil || &deployment.Spec.Template == nil || &deployment.Spec.Template.Spec == nil || &deployment.Spec.Template.Spec.Containers == nil || len(deployment.Spec.Template.Spec.Containers) == 0 {
		return
	}

	for idx, spec := range deployment.Spec.Template.Spec.Containers {
		if spec.Name == name {
			deployment.Spec.Template.Spec.Containers[idx].Image = image
		}
	}
}
