package controller

import (
	"fmt"
	"github.com/kubernetes-misc/kudecs/client"
	"github.com/kubernetes-misc/kudecs/model"
	"github.com/sirupsen/logrus"
)

var ReconHub = NewReconHub()

func NewReconHub() *reconHub {
	r := &reconHub{
		in:     make(chan model.KudecsV1, 256),
		delete: make(chan model.KudecsV1, 256),
	}
	go func() {
		for {
			select {
			case cs := <-r.in:
				logrus.Debugln("recon hub has received", cs.GetID(), "event")
				checkAndUpdate(cs)
			case cs := <-r.delete:
				logrus.Debugln("recon hub has received", cs.GetID(), "event (delete)")
				deleteCerts(cs)
			}
		}

	}()
	return r
}

type reconHub struct {
	in     chan model.KudecsV1
	delete chan model.KudecsV1
}

func (r *reconHub) Add(cs model.KudecsV1) {
	r.in <- cs
}
func (r *reconHub) Remove(cs model.KudecsV1) {
	r.delete <- cs
}

func checkAndUpdate(cs model.KudecsV1) {
	reconcileMasterKudec(cs)
	reconcileInjected(cs)
}

func deleteCerts(cs model.KudecsV1) {
	//TODO: figure out if we should remove and how to remove the injected entries
	logrus.Println(fmt.Sprintf("> Removing kudec %s/%s derived objects", cs.Metadata.Namespace, cs.Metadata.Name))

	//For now delete the master
	err := client.DeleteSecret(model.StoreNamespace, cs.GetMasterSecretName())
	if err != nil {
		logrus.Errorln(fmt.Sprintf("could not delete master secrets %s/%s", model.StoreNamespace, cs.GetMasterSecretName()))
		logrus.Println(model.LogFAIL)
		return
	}
	logrus.Println(model.LogOK)

}
