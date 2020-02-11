package model

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var KudecsV1CRDSchema = schema.GroupVersionResource{
	Group:    "kubernetes-misc.xyz",
	Version:  "v1",
	Resource: "kudecs",
}

type KudecsV1 struct {
	Metadata MetadataV1 `json:"metadata"`
	Spec     SpecV1     `json:"spec"`
}

func (cs KudecsV1) GetID() string {
	return "kudecsV1." + cs.Metadata.Namespace + "." + cs.Metadata.Name
}

func (cs KudecsV1) GetMasterSecretName() string {
	return "kudecs-v1-master-" + cs.Metadata.Namespace + "-" + cs.Metadata.Name
}

type MetadataV1 struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type SpecV1 struct {
	Days               int                 `json:"days"`
	CountryName        string              `json:"countryName"`
	StateName          string              `json:"stateName"`
	LocalityName       string              `json:"localityName"`
	OrganizationName   string              `json:"organizationName"`
	OrganizationalUnit string              `json:"organizationalUnit"`
	CommonName         string              `json:"commonName"`
	EmailAddress       string              `json:"emailAddress"`
	InjectedSecrets    []InjectedSecretsV1 `json:"injectedSecrets"`
}

type InjectedSecretsV1 struct {
	Namespace  string `json:"namespace"`
	SecretName string `json:"secretName"`
	SourceKey  string `json:"sourceKey"`
	KeyName    string `json:"keyName"`
}
