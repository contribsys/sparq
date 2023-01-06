package util

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSnowflake(t *testing.T) {
	t.Parallel()

	sgen := NewSnowflake()

	sid := sgen.NextID()
	sid2 := sgen.NextID()
	assert.Greater(t, sid2, sid)

	ssid := fmt.Sprintf("%d", sid)
	assert.Greater(t, len(ssid), 16)

}
func TestCompression(t *testing.T) {
	t.Parallel()

	sgen := NewSnowflake()

	sid := sgen.NextSID()
	sid2 := sgen.NextSID()
	assert.Greater(t, sid2, sid)

	assert.Equal(t, "AAAAAAAAAAAD9", encode(127))
}

func BenchmarkSnowflake(b *testing.B) {
	sgen := NewSnowflake()
	for idx := 0; idx < b.N; idx++ {
		sgen.NextID()
	}
}

func BenchmarkSID(b *testing.B) {
	sgen := NewSnowflake()
	for idx := 0; idx < b.N; idx++ {
		sgen.NextSID()
	}
}
