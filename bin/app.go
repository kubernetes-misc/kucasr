package main

import (
	"github.com/kubernetes-misc/kudecs/client"
	"github.com/kubernetes-misc/kudecs/controller"
	"github.com/kubernetes-misc/kudecs/model"
	cronV3 "github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"os"
)

const DefaultCronSpec = "*/30 * * * * *"
const Version = "v1.1.2"

func main() {
	logrus.Println("Kubernetes Declarative Certificates Secrets")
	logrus.Println("Starting up...")
	logrus.Println(Version)

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
	diff := client.WatchCRDS(model.KudecsV1CRDSchema)
	go func() {
		for d := range diff {
			if d.Type == "DELETED" {
				logrus.Debugln("deleted...")
				controller.ReconHub.Remove(d.Object)
			} else {
				logrus.Debugln("created...")
				controller.ReconHub.Add(d.Object)
			}
		}
	}()
	select {}

}

func update() {
	logrus.Debugln(">> Getting CRDs in all namespaces")
	crds, err := client.GetAllCRD("", model.KudecsV1CRDSchema)
	if err != nil {
		logrus.Errorln(err)
		return
	}
	for _, a := range crds {
		controller.ReconHub.Add(a)
	}

}
