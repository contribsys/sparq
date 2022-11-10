package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	t.Parallel()

	pub, priv := GenerateKeys()
	assert.NotNil(t, pub)
	assert.NotNil(t, priv)
}
