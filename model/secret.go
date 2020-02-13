package model

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewInjectSecret(c InjectedSecretsV1, masterSecret *v1.Secret) *v1.Secret {
	s := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: c.SecretName,
			Labels: map[string]string{
				ExpiresLabel + "-" + c.KeyName: masterSecret.Labels[ExpiresLabel],
			},
		},
		Data: map[string][]byte{
			c.KeyName: masterSecret.Data[c.SourceKey],
		},
		Type: v1.SecretTypeOpaque,
	}
	return s
}
