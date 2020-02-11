package controller

import (
	"fmt"
	"github.com/kubernetes-misc/kudecs/client"
	"github.com/kubernetes-misc/kudecs/gen"
	"github.com/kubernetes-misc/kudecs/model"
	"github.com/sirupsen/logrus"
	cv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"time"
)

func reconcileMaster(cs model.KudecsV1) {
	/////////////////////////////////////////////////
	/*
		WHAT NEEDS TO BE DONE
	*/
	/////////////////////////////////////////////////
	create, update := getMasterSecretTasks(cs)
	logrus.Debugln(fmt.Sprintf("getMasterTasks returning create: %t, update: %t", create, update))

	/////////////////////////////////////////////////
	/*
		SOURCE OF MASTER SECRET
	*/
	/////////////////////////////////////////////////
	var masterSecret *cv1.Secret
	if !create {
		var err error
		masterSecret, err = client.GetSecret(model.StoreNamespace, cs.GetMasterSecretName())
		if err != nil {
			logrus.Errorln("unexpected error! could not get master secret when create == false. Should not happen. Skipping")
			return
		}
	}

	/////////////////////////////////////////////////
	/*
		MODIFYING THE SECRET
	*/
	/////////////////////////////////////////////////
	if create {
		logrus.Infoln("> Generating master certificate")
		logrus.Infoln(fmt.Sprintf("  Requester: %s/%s", cs.Metadata.Namespace, cs.Metadata.Name))
		logrus.Infoln(fmt.Sprintf("  Master stored as: %s/%s", model.StoreNamespace, cs.GetMasterSecretName()))
		masterSecret = newMasterSecret(cs)
	}
	if update {
		logrus.Infoln("> Updating master certificate")
		logrus.Infoln(fmt.Sprintf("  Requester: %s/%s", cs.Metadata.Namespace, cs.Metadata.Name))
		logrus.Infoln(fmt.Sprintf("  Master stored as: %s/%s", model.StoreNamespace, cs.GetMasterSecretName()))
		updateMasterSecret(cs, masterSecret)
	}

	/////////////////////////////////////////////////
	/*
		PERSISTING THE SECRET
	*/
	/////////////////////////////////////////////////
	if create {
		err := client.CreateSecret(model.StoreNamespace, masterSecret)
		if err != nil {
			logrus.Errorln(model.LogFAIL)
			logrus.Errorln(fmt.Sprintf("  could not generate master certificate: %s/%s", model.StoreNamespace, masterSecret.Name))
			return
		}
		logrus.Infoln(model.LogOK)
		return
	}

	if update {
		err := client.UpdateSecret(model.StoreNamespace, masterSecret)
		if err != nil {
			logrus.Errorln(model.LogFAIL)
			logrus.Errorln(fmt.Sprintf("  could not update master certificate: %s/%s", model.StoreNamespace, masterSecret.Name))
			return
		}
		logrus.Infoln(model.LogOK)
		return

	}

}

func getMasterSecretTasks(cs model.KudecsV1) (create, update bool) {
	masterSecret, err := client.GetSecret(model.StoreNamespace, cs.GetMasterSecretName())
	if err != nil || masterSecret == nil {
		//The secret was not found
		create = true
		return
	}

	expiresS, hasKey := masterSecret.Labels[model.ExpiresLabel]
	if !hasKey {
		logrus.Errorln(fmt.Sprintf("master secret (%s) does not have label: %s", cs.GetMasterSecretName(), model.ExpiresLabel))
		update = true
		return
	}
	unixNano, err := strconv.Atoi(expiresS)
	if err != nil {
		logrus.Errorln(fmt.Sprintf("master secret (%s) expires label cannot be read as int: %s", cs.GetMasterSecretName(), expiresS))
		update = true
		return
	}
	expires := time.Unix(0, int64(unixNano))
	if expires.Before(time.Now().Add(1 * time.Hour)) {
		logrus.Infoln(fmt.Sprintf("master secret (%s) has expired", cs.GetMasterSecretName()))
		update = true
		return
	}
	return
}

func newMasterSecret(cs model.KudecsV1) (secret *cv1.Secret) {
	genReq := gen.GenerateRequest{
		CountryName:        cs.Spec.CountryName,
		StateName:          cs.Spec.StateName,
		LocalityName:       cs.Spec.LocalityName,
		OrganizationName:   cs.Spec.OrganizationName,
		OrganizationalUnit: cs.Spec.OrganizationalUnit,
		Hosts:              []string{""},
		NotBefore:          time.Now(),
		NotAfter:           time.Now().Add(time.Duration(cs.Spec.Days*24) * time.Hour),
	}
	private, public := gen.GenerateCert(genReq)
	n := fmt.Sprintf("%v", genReq.NotAfter.UnixNano())
	secret = &cv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: cs.GetMasterSecretName(),
			Labels: map[string]string{
				model.ExpiresLabel: n,
			},
		},
		Data: map[string][]byte{
			model.DefaultPrivate: private,
			model.DefaultPublic:  public,
		},
		Type: cv1.SecretTypeOpaque,
	}
	return
}

func updateMasterSecret(cs model.KudecsV1, secret *cv1.Secret) {
	genReq := gen.GenerateRequest{
		CountryName:        cs.Spec.CountryName,
		StateName:          cs.Spec.StateName,
		LocalityName:       cs.Spec.LocalityName,
		OrganizationName:   cs.Spec.OrganizationName,
		OrganizationalUnit: cs.Spec.OrganizationalUnit,
		Hosts:              []string{""},
		NotBefore:          time.Now(),
		NotAfter:           time.Now().Add(time.Duration(cs.Spec.Days*24) * time.Hour),
	}
	private, public := gen.GenerateCert(genReq)
	n := fmt.Sprintf("%v", genReq.NotAfter.UnixNano())
	secret.Labels[model.ExpiresLabel] = n
	secret.Data = map[string][]byte{
		model.DefaultPrivate: private,
		model.DefaultPublic:  public,
	}

}
