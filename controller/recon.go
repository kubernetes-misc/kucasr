package controller

import (
	"bytes"
	"fmt"
	cv1 "k8s.io/api/core/v1"
	"strconv"

	"github.com/kubernetes-misc/kudecs/client"
	"github.com/kubernetes-misc/kudecs/gen"
	"github.com/kubernetes-misc/kudecs/model"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
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
	//Check that masters first
	checkMasters(cs)
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
	for _, i := range cs.Spec.InjectPrivateNamespace {
		deleteIfWrong("private", i, masterSecret)
	}
	for _, i := range cs.Spec.InjectPublicNamespace {
		deleteIfWrong("public", i, masterSecret)
	}

	//Create missing secrets
	for _, i := range cs.Spec.InjectPrivateNamespace {
		createSecretFromMaster("private", i, masterSecret)
	}
	for _, i := range cs.Spec.InjectPublicNamespace {
		createSecretFromMaster("public", i, masterSecret)
	}

}

func createSecretFromMaster(dataKey string, i model.InjectNamespaceV1, masterSecret *cv1.Secret) {
	//Escape if secret exists as we checked that in deleteIfWrong()
	if s, err := client.GetSecret(i.Namespace, i.SecretName); err == nil && s != nil {
		return
	}
	logrus.Println(fmt.Sprintf("> Injecting secret %s/%s", i.Namespace, i.SecretName))
	logrus.Println(fmt.Sprintf("  Master secret: %s/%s", model.StoreNamespace, masterSecret.Name))
	logrus.Println(fmt.Sprintf("  Injected secret: %s", dataKey))
	logrus.Println(fmt.Sprintf("  Injected as: %s/%s", i.Namespace, i.SecretName))

	k := masterSecret.Data[dataKey]
	se := &cv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: i.SecretName,
			Labels: map[string]string{
				"expires": masterSecret.Labels["expires"],
			},
		},
		Data: map[string][]byte{
			dataKey: k,
		},
		Type: cv1.SecretTypeOpaque,
	}

	err := client.CreateSecret(i.Namespace, se)
	if err != nil {
		logrus.Errorln(err)
		logrus.Errorln("failed to store injected secret as ", model.StoreNamespace, se.Name)
		return
	}
	logrus.Println("  OK")

}

func deleteIfWrong(dataKey string, i model.InjectNamespaceV1, masterSecret *cv1.Secret) {
	secret, err := client.GetSecret(i.Namespace, i.SecretName)
	if err != nil || secret == nil {
		logrus.Debugln("secret ", i.Namespace, i.SecretName, "does not exist... and that is ok")
		return
	}
	mustDelete := false

	//Check if the secrets have different expiry dates
	mustDelete = mustDelete || secret.Labels["expires"] != masterSecret.Labels["expires"]

	//Check that private or public secret is correct
	mustDelete = mustDelete || !bytes.Equal(secret.Data[dataKey], masterSecret.Data[dataKey])

	logrus.Debugln("Should we delete the injected secret? ", mustDelete)

	if mustDelete {
		logrus.Println(fmt.Sprintf("> Deleting secret %s/%s", i.Namespace, i.SecretName))
		err := client.DeleteSecret(i.Namespace, i.SecretName)
		if err != nil {
			logrus.Errorln("  failed to delete secret ", i.Namespace, i.SecretName)
		}
	}
}

func checkMasters(cs model.KudecsV1) {
	mustGenerate := true
	masterName := cs.GetMasterSecretName()
	logrus.Debugln("getting master secret", model.StoreNamespace, masterName)
	secret, err := client.GetSecret(model.StoreNamespace, masterName)
	found := err == nil && secret != nil
	if found {
		logrus.Debugln("secret was found")
		mustGenerate = false
		expiresS, hasKey := secret.Labels["expires"]
		if !hasKey {
			mustGenerate = true
		}
		unixNano, err := strconv.Atoi(expiresS)
		if err != nil {
			logrus.Errorln("could not read value to int", expiresS)
		}
		expires := time.Unix(0, int64(unixNano))
		if expires.Before(time.Now().Add(1 * time.Hour)) { //TODO: choose how long before to swap out certs!
			//TODO: delete the master secret
			mustGenerate = true
		} else {
			logrus.Debugln("not expiring soon...")
		}
	}

	if mustGenerate {
		logrus.Infoln("> Generating master certificate")
		logrus.Infoln(fmt.Sprintf("  Requester: %s/%s", cs.Metadata.Namespace, cs.Metadata.Name))
		logrus.Infoln(fmt.Sprintf("  Master stored as: %s/%s", model.StoreNamespace, cs.GetMasterSecretName()))
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
		se := &cv1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: masterName,
				Labels: map[string]string{
					"expires": n,
				},
			},
			Data: map[string][]byte{
				"private": private,
				"public":  public,
			},
			Type: cv1.SecretTypeOpaque,
		}
		err = client.CreateSecret(model.StoreNamespace, se)
		if err != nil {
			logrus.Errorln(err)
			logrus.Errorln("failed to store master secret as ", model.StoreNamespace, se.Name)
			return
		}
		logrus.Println("  OK")

	}
}
