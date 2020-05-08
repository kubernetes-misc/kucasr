package openssl

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
	"time"
)

func TestGenerate(t *testing.T) {
	private, public := Generate(GenerateRequest{
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(7 * 24 * time.Hour),
	})

	logrus.Println("caller")
	logrus.Println(string(private))

	assert.NotEmpty(t, private)
	assert.NotEmpty(t, public)

	ioutil.WriteFile("private.key", private, 0644)
	ioutil.WriteFile("public.pem", public, 0644)

}
