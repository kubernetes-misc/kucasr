package main

import (
	"encoding/json"
	"fmt"
	"github.com/kubernetes-misc/kudecs/client"
	"github.com/kubernetes-misc/kudecs/model"
	cronV3 "github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/pretty"
	"os"
	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
)

const DefaultCronSpec = "*/10 * * * * *"

func main() {
	logrus.Println("KUBERNETES DECLARATIVE CERTS AS SECRETS")
	logrus.Println("Starting up...")
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
		F: updateCronScales,
	})
	c.Start()
	updateCronScales()
	if err != nil {
		panic(err)
	}
	logrus.SetLevel(logrus.InfoLevel)
	select {}

}

func updateCronScales() {

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

	//TODO: allCRDs

	for _, a := range allCRDs {
		b, _ := json.Marshal(a)
		fmt.Println(string(pretty.Pretty(b)))
	}

}
