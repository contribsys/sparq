package clientapi

import (
	"fmt"
	"testing"

	"github.com/contribsys/sparq/model"
	"github.com/contribsys/sparq/web"
	"github.com/stretchr/testify/assert"
)

func TestLocal(t *testing.T) {
	ts, stopper := web.NewTestServer(t, "timeline")
	defer stopper()

	var count int
	err := ts.DB().QueryRow("select count(*) from toots").Scan(&count)
	assert.NoError(t, err)
	fmt.Printf("Found %d toots\n", count)

	tq := model.TQ(ts.DB())
	tq.Local = true
	// tq.only_media = true
	res, err := tq.Execute()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(res.Toots))

	tq = model.TQ(ts.DB())
	// tq.local = true
	tq.OnlyMedia = true
	res, err = tq.Execute()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(res.Toots))
}
