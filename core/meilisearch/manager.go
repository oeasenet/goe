package meilisearch

import (
	"fmt"
	"github.com/meilisearch/meilisearch-go"
	"strings"
)

// Connect initializes a MeiliSearch connection based on the provided configuration prefix.
// The configuration keys are expected to be like:
// MEILISEARCH_URI, MEILISEARCH_API_KEY
// For a named connection "foo", the keys would be:
// MeiliSearch_FOO_URI, MeiliSearch_FOO_API_KEY
func (ms *MeiliSearch) connect(name string) (meilisearch.ServiceManager, error) {
	configPrefix := "MEILISEARCH_"
	if name != "default" && name != "" {
		configPrefix = fmt.Sprintf("MEILISEARCH_%s_", strings.ToUpper(name))
	}

	// get uri and database
	uri := ms.config.GetString(configPrefix + "URI")
	apiKey := ms.config.GetString(configPrefix + "API_KEY")

	if uri == "" {
		return nil, fmt.Errorf("no URI for %s meilisearch", name)
	}
	if apiKey == "" {
		return nil, fmt.Errorf("no database name for %s meilisearch connection", name)
	}

	client := meilisearch.New(uri, meilisearch.WithAPIKey(apiKey))

	if !client.IsHealthy() {
		return nil, fmt.Errorf("could not connect to %s meilisearch", name)
	}
	ms.logger.Info("meilisearch connection established successfully",
		"name", name,
	)

	return client, nil
}
