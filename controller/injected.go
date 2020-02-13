package controller

import (
	"bytes"
	"fmt"
	"github.com/kubernetes-misc/kudecs/client"
	"github.com/kubernetes-misc/kudecs/model"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

func reconcileInjected(cs model.KudecsV1) {

	create, update := getInjectedSecretTasks(cs)
	logrus.Debugln(fmt.Sprintf("getInjectedSecretTasks returning create: %v, update: %v", len(create), len(update)))

	masterSecret, err := client.GetSecret(model.StoreNamespace, cs.GetMasterSecretName())
	if err != nil {
		logrus.Errorln(fmt.Sprintf("unexpected error getting master secret (%s/%s) during create of injected secret", model.StoreNamespace, cs.GetMasterSecretName()))
		return
	}

	reconcileInjectedCreates(cs, create, masterSecret)
	reconcileInjectedUpdates(cs, update, masterSecret)

}

func reconcileInjectedCreates(cs model.KudecsV1, create []model.InjectedSecretsV1, masterSecret *v1.Secret) {
	for _, c := range create {
		logrus.Infoln("> Creating injected certificate")
		logrus.Infoln(fmt.Sprintf("  Requester: %s/%s", cs.Metadata.Namespace, cs.Metadata.Name))
		logrus.Infoln(fmt.Sprintf("  Master stored as: %s/%s", model.StoreNamespace, cs.GetMasterSecretName()))
		logrus.Infoln(fmt.Sprintf("  Secret to be created: %s/%s", c.Namespace, c.SecretName))

		s := model.NewInjectSecret(c, masterSecret)

		err := client.CreateSecret(c.Namespace, s)
		if err != nil {
			logrus.Errorln(fmt.Sprintf("unexpected error creating injected secret (%s/%s)", c.Namespace, c.SecretName))
			logrus.Infoln(model.LogFAIL)
			continue
		}
		logrus.Infoln(model.LogOK)

	}

}

func reconcileInjectedUpdates(cs model.KudecsV1, update []model.InjectedSecretsV1, masterSecret *v1.Secret) {
	for _, c := range update {
		logrus.Infoln("> Updating injected certificate")
		logrus.Infoln(fmt.Sprintf("  Requester: %s/%s", cs.Metadata.Namespace, cs.Metadata.Name))
		logrus.Infoln(fmt.Sprintf("  Master stored as: %s/%s", model.StoreNamespace, cs.GetMasterSecretName()))
		logrus.Infoln(fmt.Sprintf("  Secret to be updated: %s/%s", c.Namespace, c.SecretName))
		s, err := client.GetSecret(c.Namespace, c.SecretName)
		if err != nil {
			logrus.Errorln(fmt.Sprintf("unexpected error getting injected secret (%s/%s)", c.Namespace, c.SecretName))
			logrus.Infoln(model.LogFAIL)
			continue
		}
		s.Labels[model.ExpiresLabel+"-"+c.KeyName] = masterSecret.Labels[model.ExpiresLabel]
		s.Data[c.KeyName] = masterSecret.Data[c.SourceKey]
		err = client.UpdateSecret(c.Namespace, s)
		if err != nil {
			logrus.Errorln(fmt.Sprintf("unexpected error updating injected secret (%s/%s)", c.Namespace, c.SecretName))
			logrus.Errorln(err)
			logrus.Infoln(model.LogFAIL)
			continue
		}
		logrus.Infoln(model.LogOK)
	}

}

func getInjectedSecretTasks(cs model.KudecsV1) (create []model.InjectedSecretsV1, update []model.InjectedSecretsV1) {
	masterSecret, err := client.GetSecret(model.StoreNamespace, cs.GetMasterSecretName())
	if err != nil || masterSecret == nil {
		logrus.Errorln(fmt.Sprintf("unexpected error. Master secret (%s/%s) should exist. Skipping injected", model.StoreNamespace, cs.GetMasterSecretName()))
		return
	}

	for _, i := range cs.Spec.InjectedSecrets {
		s, err := client.GetSecret(i.Namespace, i.SecretName)
		if err != nil || s == nil {
			logrus.Infoln(fmt.Sprintf("  Injected cert %s/%s will need to be created", i.Namespace, i.SecretName))
			create = append(create, i)
			continue
		}
		if !certsEqual(masterSecret, s, i) {
			logrus.Infoln(fmt.Sprintf("  Injected cert %s/%s will need to be updated", i.Namespace, i.SecretName))
			update = append(update, i)
		}
	}

	return
}

func certsEqual(master, secret *v1.Secret, is model.InjectedSecretsV1) bool {
	if master.Labels[model.ExpiresLabel] != secret.Labels[model.ExpiresLabel+"-"+is.KeyName] {
		return false
	}
	if !bytes.Equal(master.Data[is.SourceKey], secret.Data[is.KeyName]) {
		return false
	}
	return true
}
