package model

import (
	"testing"

	"github.com/contribsys/sparq/db"
	"github.com/stretchr/testify/assert"
)

func TestQuery(t *testing.T) {
	db, stopper, err := db.TestDB("query")
	assert.NoError(t, err)
	defer stopper()

	tq := TQ(db)
	tq.Local = true
	result, err := tq.Execute()
	assert.NoError(t, err)

	assert.NotNil(t, result)
	assert.NotNil(t, result.Toots)
	assert.EqualValues(t, 2, len(result.Toots))
	assert.True(t, !result.IsEmpty())

	for _, entry := range result.Toots {
		assert.NotEmpty(t, entry.Content)

		att, err := entry.MediaAttachments()
		assert.NoError(t, err)
		assert.NotNil(t, att)

		tags, err := entry.Tags()
		assert.NoError(t, err)
		assert.NotNil(t, tags)
	}
}
