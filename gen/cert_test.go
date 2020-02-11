package gen

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGenerateCert(t *testing.T) {
	private, public := GenerateCert(GenerateRequest{
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(7 * 24 * time.Hour),
	})

	assert.NotEmpty(t, private)
	assert.NotEmpty(t, public)

}
