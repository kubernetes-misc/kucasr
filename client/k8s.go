package client

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/kubernetes-misc/kudecs/model"
	"github.com/sirupsen/logrus"
	cv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

var clientset *kubernetes.Clientset
var dynClient dynamic.Interface

func BuildClient() (err error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		logrus.Errorln(err)
		return
	}
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		logrus.Errorln(err)
		return
	}

	dynClient, err = dynamic.NewForConfig(config)
	if err != nil {
		logrus.Errorln(fmt.Sprintf("Error received creating client %v", err))
		return
	}
	return
}

func GetAllNS() ([]string, error) {
	logrus.Debugln("== getting namespaces ==")
	ls, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		logrus.Errorln(err)
		return nil, err
	}
	result := make([]string, len(ls.Items))
	for i, n := range ls.Items {
		result[i] = n.Name
	}
	return result, nil
}

func GetAllCRD(namespace string, crd schema.GroupVersionResource) (result []model.KudecsV1, err error) {
	logrus.Debugln("== getting CRDs ==")
	crdClient := dynClient.Resource(crd)
	ls, err := crdClient.Namespace(namespace).List(metav1.ListOptions{})
	if err != nil {
		logrus.Errorln(fmt.Errorf("could not list %s", err))
		return
	}

	result = make([]model.KudecsV1, len(ls.Items))
	for i, entry := range ls.Items {
		b, err := entry.MarshalJSON()
		if err != nil {
			logrus.Errorln(err)
			continue
		}
		cs := model.KudecsV1{}
		err = json.Unmarshal(b, &cs)
		if err != nil {
			logrus.Errorln(err)
		}
		result[i] = cs
	}
	return
}

func GetSecret(ns, name string) (secret *cv1.Secret, err error) {
	secret, err = clientset.CoreV1().Secrets(ns).Get(name, metav1.GetOptions{})
	return
}

func CreateSecret(ns string, secret *cv1.Secret) error {
	_, err := clientset.CoreV1().Secrets(ns).Create(secret)
	return err
}

func DeleteSecret(ns, name string) (err error) {
	return clientset.CoreV1().Secrets(ns).Delete(name, &metav1.DeleteOptions{})
}
