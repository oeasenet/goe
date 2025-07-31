package meilisearch

import (
	"fmt"
	"github.com/meilisearch/meilisearch-go"
)

// IndexAttributeConfig defines configurable attributes for a Meilisearch index.
type IndexAttributeConfig struct {
	SearchableFields []string `json:"searchable_fields"`
	FilterableFields []string `json:"filterable_fields"`
	SortableFields   []string `json:"sortable_fields"`
	DisplayedFields  []string `json:"displayed_fields"`
}

type ConfigDataMap map[string]*IndexAttributeConfig

// ApplyIndexConfig applies the attribute configuration for multiple Meilisearch indexes.
func (ms *MeiliSearch) ApplyIndexConfig(configMap ConfigDataMap) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	for indexName, cfg := range configMap {
		indexResult, err := ms.Instance().GetIndex(indexName)
		if err != nil {
			ms.logger.Debug("index '" + indexName + "' not exists, create it")
			_, err = ms.Instance().CreateIndex(&meilisearch.IndexConfig{
				Uid:        indexName,
				PrimaryKey: "id",
			})
		}

		// set searchableFields
		if len(cfg.SearchableFields) > 0 {
			_, err := indexResult.IndexManager.UpdateSearchableAttributes(&cfg.SearchableFields)
			if err != nil {
				return fmt.Errorf("failed to set searchable fields for index %s: %w", indexName, err)
			}
		}

		// set filterableFields
		if len(cfg.FilterableFields) > 0 {
			_, err := indexResult.IndexManager.UpdateFilterableAttributes(&cfg.FilterableFields)
			if err != nil {
				return fmt.Errorf("failed to set filterable fields for index %s: %w", indexName, err)
			}
		}

		// set sortableFields
		if len(cfg.SortableFields) > 0 {
			_, err := indexResult.IndexManager.UpdateSortableAttributes(&cfg.SortableFields)
			if err != nil {
				return fmt.Errorf("failed to set sortable fields for index %s: %w", indexName, err)
			}
		}

		// set displayedFields
		if len(cfg.DisplayedFields) > 0 {
			_, err := indexResult.IndexManager.UpdateDisplayedAttributes(&cfg.DisplayedFields)
			if err != nil {
				return fmt.Errorf("failed to set displayed fields for index %s: %w", indexName, err)
			}
		}
	}

	return nil
}
