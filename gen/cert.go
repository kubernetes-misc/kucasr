package gen

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"github.com/kubernetes-misc/kudecs/model"
	"log"
	"math/big"
	"net"
	"time"
)

var (
	isCA       = false
	rsaBits    = 2048
	ecdsaCurve = "P256"
	ed25519Key = false
)

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	case ed25519.PrivateKey:
		return k.Public().(ed25519.PublicKey)
	default:
		return nil
	}
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

func generatePrivate() (interface{}, error) {
	var priv interface{}
	var err error
	switch ecdsaCurve {
	case "":
		if ed25519Key {
			_, priv, err = ed25519.GenerateKey(rand.Reader)
		} else {
			priv, err = rsa.GenerateKey(rand.Reader, rsaBits)
		}
	case "P224":
		priv, err = ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	case "P256":
		priv, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case "P384":
		priv, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case "P521":
		priv, err = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	default:
		log.Fatalf("Unrecognized elliptic curve: %q", ecdsaCurve)
	}
	return priv, err
}

func GenerateCert(request GenerateRequest) (private, public []byte) {
	priv, err := generatePrivate()
	if err != nil {
		log.Fatalf("Failed to generate private key: %s", err)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("Failed to generate serial number: %s", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Country:            []string{request.CountryName},
			Organization:       []string{request.OrganizationName},
			OrganizationalUnit: []string{request.OrganizationalUnit},
			Locality:           []string{request.LocalityName},
			Province:           []string{request.StateName},
		},
		NotBefore: request.NotBefore,
		NotAfter:  request.NotAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	for _, h := range request.Hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	if isCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)
	if err != nil {
		log.Fatalf("Failed to create certificate: %s", err)
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		log.Fatalf("Unable to marshal private key: %v", err)
	}

	private = privBytes
	public = derBytes

	return

}
