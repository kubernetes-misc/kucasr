package controller

import (
	"fmt"
	"github.com/kubernetes-misc/kudecs/client"
	"github.com/kubernetes-misc/kudecs/model"
	"github.com/sirupsen/logrus"
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
	checkAndUpdateCert(cs)

}

func checkAndUpdateCert(cs model.KudecsV1) {

}

func checkAndUpdateHPA(cs model.KudecsV1) {
	//Check the hpa
	hpa, err := client.GetHPA(cs.Metadata.Namespace, cs.Spec.HorizontalPodAutoScaler.Name)
	if err != nil {
		logrus.Errorln(err)
		return
	}
	hpa.Spec.MinReplicas = &cs.Spec.HorizontalPodAutoScaler.MinReplicas
	hpa.Spec.MaxReplicas = cs.Spec.HorizontalPodAutoScaler.MaxReplicas
	hpa.Spec.TargetCPUUtilizationPercentage = &cs.Spec.HorizontalPodAutoScaler.TargetCPUUtilizationPercentage
	client.UpdateHPA(cs.Metadata.Namespace, hpa)
	logrus.Infoln(fmt.Sprintf(">> Updating hpa/%s from %v to %v @ CPU load of %v%%", cs.Spec.HorizontalPodAutoScaler.Name, cs.Spec.HorizontalPodAutoScaler.MinReplicas, cs.Spec.HorizontalPodAutoScaler.MaxReplicas, cs.Spec.HorizontalPodAutoScaler.TargetCPUUtilizationPercentage))

}
