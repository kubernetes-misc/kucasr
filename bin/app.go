package main

import (
	"github.com/kubernetes-misc/kudecs/client"
	"github.com/kubernetes-misc/kudecs/controller"
	"github.com/kubernetes-misc/kudecs/model"
	cronV3 "github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"os"
	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
)

const DefaultCronSpec = "*/10 * * * * *"

func main() {
	logrus.Println("Kubernetes Declarative Certificates Secrets")
	logrus.Println("Starting up...")

	model.StoreNamespace = os.Getenv("storeNamespace")
	if model.StoreNamespace == "" {
		logrus.Fatalln("no storeNamespace set! You must provide a namespace in which to store the master copy of the secrets as storeNamespace env variable")
	}
	err := client.BuildClient()
	if err != nil {
		panic(err)
	}
	cronSpec := os.Getenv("cronSpec")
	if cronSpec == "" {
		logrus.Println("cronSpec env is empty. Defaulting to", DefaultCronSpec)
		cronSpec = DefaultCronSpec
	}
	c := cronV3.New(cronV3.WithSeconds())
	_, err = c.AddJob(cronSpec, model.Job{
		F: update,
	})
	c.Start()
	update()
	if err != nil {
		panic(err)
	}
	logrus.SetLevel(logrus.InfoLevel)
	go client.WatchCRDS(model.KudecsV1CRDSchema)
	select {}

}

func update() {

	logrus.Debugln("> Getting all namespaces")
	nsl, err := client.GetAllNS()
	if err != nil {
		logrus.Errorln(err)
		return
	}

	allCRDs := make([]model.KudecsV1, 0)
	for _, ns := range nsl {
		logrus.Debugln(">> Getting CRDs in", ns)
		crds, err := client.GetAllCRD(ns, model.KudecsV1CRDSchema)
		if err != nil {
			logrus.Errorln(err)
			return
		}
		allCRDs = append(allCRDs, crds...)
	}

	for _, a := range allCRDs {
		controller.ReconHub.Add(a)
	}

}
