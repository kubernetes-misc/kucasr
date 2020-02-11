package controller

import (
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

	//checkAndUpdateCert(cs)

}

func checkMasters(cs model.KudecsV1) {
	mustGenerate := true
	masterName := cs.GetMasterSecretName()
	logrus.Println("getting secret ", model.StoreNamespace, masterName)
	secret, err := client.GetSecret(model.StoreNamespace, masterName)
	found := err == nil && secret != nil
	if found {
		logrus.Println("secret was found")
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
			logrus.Println("not expiring soon...")
		}
	}

	if mustGenerate {
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
		}

	}
}

//func checkAndUpdateCert(cs model.KudecsV1) {
//	//Check the cert exists
//	certs := make([]model.InjectNamespaceV1, 0)
//	certs = append(certs, cs.Spec.InjectPublicNamespace...)
//	certs = append(certs, cs.Spec.InjectPrivateNamespace...)
//
//	masterName := cs.GetMasterSecretName()
//	//Does the master exist?
//	secret, err := client.GetCert(cs.Metadata.Namespace, cs.Metadata.Name)
//	generateMaster := true
//	found := err != nil && secret != nil
//	if found {
//		generateMaster = false
//		//TODO: check if expired and regen?
//		if false {
//			generateMaster = true
//		}
//	}
//
//	//Generate cert
//	if generateMaster {
//		genReq := gen.GenerateRequest{
//			CountryName:        cs.Spec.CountryName,
//			StateName:          cs.Spec.StateName,
//			LocalityName:       cs.Spec.LocalityName,
//			OrganizationName:   cs.Spec.OrganizationName,
//			OrganizationalUnit: cs.Spec.OrganizationalUnit,
//			Hosts:              []string{""},
//			NotBefore:          time.Now(),
//			NotAfter:           time.Now().Add(time.Duration(cs.Spec.Days*24) * time.Hour),
//		}
//		private, public := gen.GenerateCert(genReq)
//
//		se := &cv1.Secret{
//			ObjectMeta: metav1.ObjectMeta{
//				Name:      masterName,
//				Namespace: model.StoreNamespace,
//				Labels: map[string]string{
//					"expires": string(genReq.NotAfter.UnixNano()),
//				},
//			},
//			Data: map[string][]byte{
//				"private": private,
//				"public":  public,
//			},
//			Type: cv1.SecretTypeOpaque,
//		}
//
//	}

//}
