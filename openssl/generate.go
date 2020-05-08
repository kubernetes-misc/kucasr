package openssl

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"github.com/kubernetes-misc/kudecs/model"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"
)

func Generate(request GenerateRequest) ([]byte, []byte) {

	requestID := uuid.New().String()
	publicFile := fmt.Sprint("/tmp/", requestID, ".public")
	privateFile := fmt.Sprint("/tmp/", requestID, ".private")

	//Private key
	app := "openssl"
	arg0 := "genrsa"
	arg1 := "-out"
	arg2 := privateFile
	arg3 := "2048"
	logrus.Infoln(app, arg0, arg1, arg2, arg3)
	cmd := exec.Command(app, arg0, arg1, arg2, arg3)
	stdout, err := cmd.Output()
	if err != nil {
		logrus.Errorln("could not execute openssl (private)", err)
		return nil, nil
	}
	logrus.Println((string(stdout)))

	//Public key
	arg0 = "rsa"
	arg1 = "-in"
	arg2 = privateFile
	arg3 = "-out"
	arg4 := publicFile
	arg5 := "-pubout"
	arg6 := "-outform"
	arg7 := "PEM"
	logrus.Infoln(app, arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7)
	cmd = exec.Command(app, arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7)
	stdout, err = cmd.Output()
	if err != nil {
		logrus.Errorln("could not execute openssl (public)", err)
		return nil, nil
	}
	log.Println(string(stdout))

	/// openssl pkcs8 -topk8 -inform PEM -outform DER -in privkey.pem -nocrypt > pkcs8_key
	time.Sleep(1 * time.Second)
	arg0 = "pkcs8"
	arg1 = "-topk8"
	arg2 = "-inform"
	arg3 = "PEM"
	arg4 = "-outform"
	arg5 = "DER"
	arg6 = "-in"
	arg7 = privateFile
	arg8 := "-nocrypt"
	logrus.Infoln(app, arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8)
	cmd = exec.Command(app, arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8)

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err = cmd.Run()
	if err != nil {
		logrus.Errorln(err)
		logrus.Errorln("out:", outb.String(), "err:", errb.String())
	}
	private := outb.Bytes()

	public, err := ioutil.ReadFile(publicFile)

	os.Remove(publicFile)
	os.Remove(privateFile)

	return private, public
}

func NewGenerateRequest(cs model.KudecsV1) GenerateRequest {
	result := GenerateRequest{
		CountryName:        cs.Spec.CountryName,
		StateName:          cs.Spec.StateName,
		LocalityName:       cs.Spec.LocalityName,
		OrganizationName:   cs.Spec.OrganizationName,
		OrganizationalUnit: cs.Spec.OrganizationalUnit,
		Hosts:              []string{""},
		NotBefore:          time.Now(),
		NotAfter:           time.Now().Add(time.Duration(cs.Spec.Days*24) * time.Hour),
	}
	return result
}

type GenerateRequest struct {
	CountryName        string
	StateName          string
	LocalityName       string
	OrganizationName   string
	OrganizationalUnit string
	Hosts              []string
	NotBefore          time.Time
	NotAfter           time.Time
}
