package msearch

import (
	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewMSearch(t *testing.T) {
	search := NewMSearch("http://localhost:7700", "")
	search.client.IsHealthy()
	assert.True(t, search.client.IsHealthy())
}

func TestApplyIndexConfigs(t *testing.T) {
	dataMap := configDataMap{}
	dataMap["movies"] = &IndexAttributeItem{
		SearchableFields: []string{"title"},
		FilterableFields: []string{"id"},
		SortableFields:   []string{"release_date"},
		DisplayedFields:  []string{"id", "title", "overview", "genres", "poster", "release_date"},
	}
	data, _ := json.Marshal(dataMap)
	search := NewMSearch("http://localhost:7700", "")
	err := search.ApplyIndexConfigs(data)
	assert.NoError(t, err)

	index, err := search.client.GetIndex("movies")
	assert.NoError(t, err)

	resp, err := index.GetSearchableAttributes()
	assert.NoError(t, err)
	assert.Equal(t, resp, &[]string{"title"})

	resp, err = index.GetFilterableAttributes()
	assert.NoError(t, err)
	assert.Equal(t, resp, &[]string{"id"})

	resp, err = index.GetSortableAttributes()
	assert.NoError(t, err)
	assert.Equal(t, resp, &[]string{"release_date"})

	resp, err = index.GetDisplayedAttributes()
	assert.NoError(t, err)
	assert.Equal(t, resp, &[]string{"id", "title", "overview", "genres", "poster", "release_date"})
}
