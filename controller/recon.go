package controller

import (
	"bytes"
	"fmt"
	"github.com/kubernetes-misc/kudecs/client"
	"github.com/kubernetes-misc/kudecs/model"
	"github.com/sirupsen/logrus"
	cv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ReconHub = NewReconHub()

func NewReconHub() *reconHub {
	r := &reconHub{in: make(chan model.KudecsV1, 256)}
	go func() {
		for cs := range r.in {
			logrus.Debugln("recon hub has received", cs.GetID(), "event")
			checkAndUpdate(cs)
		}
	}()
	return r
}

type reconHub struct {
	in chan model.KudecsV1
}

func (r *reconHub) Add(cs model.KudecsV1) {
	r.in <- cs
}

func checkAndUpdate(cs model.KudecsV1) {
	reconcileMaster(cs)
	checkInjected(cs)
}

func checkInjected(cs model.KudecsV1) {
	//Get the master
	masterSecret, err := client.GetSecret(model.StoreNamespace, cs.GetMasterSecretName())
	if err != nil || masterSecret == nil {
		logrus.Errorln("checkInjected could not find the secret!")
		return
	}

	//Delete broken / out of data secrets
	for _, i := range cs.Spec.InjectedSecrets {
		deleteIfWrong(i, masterSecret)
	}

	//Create missing secrets
	for _, i := range cs.Spec.InjectedSecrets {
		createSecretFromMaster(i, masterSecret)
	}

}

func createSecretFromMaster(i model.InjectedSecretsV1, masterSecret *cv1.Secret) {
	//Escape if secret exists as we checked that in deleteIfWrong()
	if s, err := client.GetSecret(i.Namespace, i.SecretName); err == nil && s != nil {
		return
	}
	logrus.Println(fmt.Sprintf("> Injecting secret %s/%s", i.Namespace, i.SecretName))
	logrus.Println(fmt.Sprintf("  Master secret: %s/%s", model.StoreNamespace, masterSecret.Name))
	logrus.Println(fmt.Sprintf("  K8s object: %s/%s", i.Namespace, i.SecretName))
	logrus.Println(fmt.Sprintf("  Injected from: %s", i.SourceKey))
	logrus.Println(fmt.Sprintf("  Injected as: %s", i.KeyName))

	k := masterSecret.Data[i.SourceKey]
	se := &cv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: i.SecretName,
			Labels: map[string]string{
				model.ExpiresLabel: masterSecret.Labels[model.ExpiresLabel],
			},
		},
		Data: map[string][]byte{
			i.KeyName: k,
		},
		Type: cv1.SecretTypeOpaque,
	}

	err := client.CreateSecret(i.Namespace, se)
	if err != nil {
		logrus.Errorln(err)
		logrus.Errorln("failed to store injected secret as ", model.StoreNamespace, se.Name)
		return
	}
	logrus.Println(model.LogOK)

}

func deleteIfWrong(i model.InjectedSecretsV1, masterSecret *cv1.Secret) {
	secret, err := client.GetSecret(i.Namespace, i.SecretName)
	if err != nil || secret == nil {
		logrus.Debugln("secret ", i.Namespace, i.SecretName, "does not exist... and that is ok")
		return
	}
	mustDelete := false

	//Check if the secrets have different expiry dates
	mustDelete = mustDelete || secret.Labels[model.ExpiresLabel] != masterSecret.Labels[model.ExpiresLabel]

	//Check that private or public secret is correct
	mustDelete = mustDelete || !bytes.Equal(secret.Data[i.KeyName], masterSecret.Data[i.SourceKey])

	logrus.Debugln("Should we delete the injected secret? ", mustDelete, i.SecretName)

	if mustDelete {
		logrus.Println(fmt.Sprintf("> Deleting secret %s/%s", i.Namespace, i.SecretName))
		err := client.DeleteSecret(i.Namespace, i.SecretName)
		if err != nil {
			logrus.Errorln("  failed to delete secret ", i.Namespace, i.SecretName)
		}
	}
}
