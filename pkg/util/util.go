package util

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"
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

func NameWithNamespace(namespace, name, suffix string) string {
	return fmt.Sprintf("%s-%s-%s", namespace, name, suffix)
}
