package clientapi

import (
	"testing"

	"github.com/contribsys/sparq/db"
	"github.com/stretchr/testify/assert"
)

func TestLocal(t *testing.T) {
	stopper, err := db.TestDB("status")
	assert.NoError(t, err)
	defer stopper()

	tq := TQ(db.Database())
	tq.local = true
	res, err := tq.Execute()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(res))
}
