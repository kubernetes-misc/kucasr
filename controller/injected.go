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

	/////////////////////////////////////////////////
	/*
		WHAT NEEDS TO BE DONE
	*/
	/////////////////////////////////////////////////
	//create, update, merge := getInjectedSecretTasks(cs)
	//logrus.Debugln(fmt.Sprintf("getInjectedSecretTasks returning create: %v, update: %v, merge: %v", len(create), len(update), len(merge)))
	//inject := make(map[string]*v1.Secret)

	/////////////////////////////////////////////////
	/*
		SOURCE OF MASTER SECRET
	*/
	/////////////////////////////////////////////////
	//for _, c := range create {
	//	masterSecret, err := client.GetSecret(model.StoreNamespace, cs.GetMasterSecretName())
	//	if err != nil {
	//		logrus.Errorln(fmt.Sprintf("unexpected error getting master secret (%s/%s) during create of injected secret", model.StoreNamespace, cs.GetMasterSecretName()))
	//		continue
	//	}
	//
	//	k := masterSecret.Data[c]
	//	se := &cv1.Secret{
	//		ObjectMeta: metav1.ObjectMeta{
	//			Name: i.SecretName,
	//			Labels: map[string]string{
	//				model.ExpiresLabel: masterSecret.Labels[model.ExpiresLabel],
	//			},
	//		},
	//		Data: map[string][]byte{
	//			dataKey: k,
	//		},
	//		Type: cv1.SecretTypeOpaque,
	//	}
	//}
	//var masterSecret *cv1.Secret
	//if !create {
	//	var err error
	//	masterSecret, err = client.GetSecret(model.StoreNamespace, cs.GetMasterSecretName())
	//	if err != nil {
	//		logrus.Errorln("unexpected error! could not get master secret when create == false. Should not happen. Skipping")
	//		return
	//	}
	//}
	//
	///////////////////////////////////////////////////
	///*
	//	MODIFYING THE SECRET
	//*/
	///////////////////////////////////////////////////
	//if create {
	//	logrus.Infoln("> Generating master certificate")
	//	logrus.Infoln(fmt.Sprintf("  Requester: %s/%s", cs.Metadata.Namespace, cs.Metadata.Name))
	//	logrus.Infoln(fmt.Sprintf("  Master stored as: %s/%s", model.StoreNamespace, cs.GetMasterSecretName()))
	//	masterSecret = newMasterSecret(cs)
	//}
	//if update {
	//	logrus.Infoln("> Updating master certificate")
	//	logrus.Infoln(fmt.Sprintf("  Requester: %s/%s", cs.Metadata.Namespace, cs.Metadata.Name))
	//	logrus.Infoln(fmt.Sprintf("  Master stored as: %s/%s", model.StoreNamespace, cs.GetMasterSecretName()))
	//	updateMasterSecret(cs, masterSecret)
	//}
	//
	///////////////////////////////////////////////////
	///*
	//	PERSISTING THE SECRET
	//*/
	///////////////////////////////////////////////////
	//if create {
	//	err := client.CreateSecret(model.StoreNamespace, masterSecret)
	//	if err != nil {
	//		logrus.Errorln(model.LogFAIL)
	//		logrus.Errorln(fmt.Sprintf("  could not generate master certificate: %s/%s", model.StoreNamespace, masterSecret.Name))
	//		return
	//	}
	//	logrus.Infoln(model.LogOK)
	//	return
	//}
	//
	//if update {
	//	err := client.UpdateSecret(model.StoreNamespace, masterSecret)
	//	if err != nil {
	//		logrus.Errorln(model.LogFAIL)
	//		logrus.Errorln(fmt.Sprintf("  could not update master certificate: %s/%s", model.StoreNamespace, masterSecret.Name))
	//		return
	//	}
	//	logrus.Infoln(model.LogOK)
	//	return
	//
	//}

}

func getInjectedSecretTasks(cs model.KudecsV1) (create []model.InjectNamespaceV1, update []model.InjectNamespaceV1, merge []model.InjectNamespaceV1) {
	masterSecret, err := client.GetSecret(model.StoreNamespace, cs.GetMasterSecretName())
	if err != nil || masterSecret == nil {
		logrus.Errorln(fmt.Sprintf("unexpected error. Master secret (%s/%s) should exist. Skipping injected", model.StoreNamespace, cs.GetMasterSecretName()))
		return
	}

	for _, i := range cs.Spec.InjectPrivateNamespace {
		s, err := client.GetSecret(i.Namespace, i.SecretName)
		if err != nil || s == nil {
			create = append(create, i)
			continue
		}
		if !certsEqual(masterSecret, s, model.DefaultPrivate) {
			update = append(update, i)
		}
	}
	for _, i := range cs.Spec.InjectPublicNamespace {
		s, err := client.GetSecret(i.Namespace, i.SecretName)
		if err != nil || s == nil {
			create = append(create, i)
			continue
		}
		if !certsEqual(masterSecret, s, model.DefaultPublic) {
			update = append(update, i)
		}
	}
	return
}

func certsEqual(master, secret *v1.Secret, dataKey string) bool {
	if master.Labels[model.ExpiresLabel] != secret.Labels[model.ExpiresLabel] {
		return false
	}
	if !bytes.Equal(master.Data[dataKey], secret.Data[dataKey]) {
		return false
	}
	return true
}
