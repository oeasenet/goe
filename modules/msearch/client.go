package msearch

import (
	"github.com/goccy/go-json"
	"github.com/meilisearch/meilisearch-go"
	"sync"
)

type MSearch struct {
	initialized bool
	client      meilisearch.ServiceManager
	once        sync.Once
	indexConfig *IndexConfigs
	logger      Logger
}

func NewMSearch(hostUrl string, key string, logger ...Logger) *MSearch {
	ms := &MSearch{}
	ms.client = meilisearch.New(hostUrl, meilisearch.WithAPIKey(key))
	if len(logger) > 0 && logger[0] != nil {
		ms.logger = logger[0]
	} else {
		ms.logger = newDefaultLogger()
	}
	return ms
}

func (ms *MSearch) ApplyIndexConfigs(configData []byte) error {
	cfg := &IndexConfigs{}
	cfg.ConfigData = &configDataMap{}
	err := json.Unmarshal(configData, cfg.ConfigData)
	if err != nil {
		return err
	}
	ms.indexConfig = cfg
	ms.initialized = true
	ms.once.Do(func() {
		//taskIDs := make([]int64, len(*ms.indexConfig.ConfigData))
		//cnt := 0
		for indexName, indexConfig := range *ms.indexConfig.ConfigData {
			// check if index exists
			_, err := ms.client.GetIndex(indexName)
			if err != nil {
				ms.logger.Debug("index '" + indexName + "' not exists, create it")
				// maybe index not exists, create it
				_, err = ms.client.CreateIndex(&meilisearch.IndexConfig{
					Uid:        indexName,
					PrimaryKey: "id",
				})
			}
			// set index attributes
			_, err = ms.client.Index(indexName).UpdateSettings(&meilisearch.Settings{
				SearchableAttributes: indexConfig.SearchableFields,
				FilterableAttributes: indexConfig.FilterableFields,
				SortableAttributes:   indexConfig.SortableFields,
				DisplayedAttributes:  indexConfig.DisplayedFields,
			})
			if err != nil {
				ms.logger.Error(err)
			}
		}
		//// wait for all tasks to finish
		//for _, taskID := range taskIDs {
		//	err := ms.WaitForTaskSuccess(taskID)
		//	if err != nil {
		//		ms.logger.Error(err)
		//	}
		//}
	})
	return nil
}
