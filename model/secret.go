package model

import (
	"fmt"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"time"
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

func GetExpiresFromSecret(secret *v1.Secret, labelName string) (expires time.Time, err error) {
	expiresS, hasKey := secret.Labels[labelName]
	if !hasKey {
		logrus.Errorln(fmt.Sprintf("secret (%s/%s) does not have label: %s", secret.Namespace, secret.Name, labelName))
		return
	}
	unixNano, err := strconv.Atoi(expiresS)
	if err != nil {
		logrus.Errorln(fmt.Sprintf("master secret (%s/%s) expires label cannot be read as int: %s", secret.Namespace, secret.Name, expiresS))
		return
	}
	expires = time.Unix(0, int64(unixNano))
	return

}
