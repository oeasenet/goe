package msearch

import (
	"context"
	"github.com/goccy/go-json"
	"github.com/meilisearch/meilisearch-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.oease.dev/omgo"
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
		for indexName, indexConfig := range *ms.indexConfig.ConfigData {
			_, err := ms.client.DeleteIndex(indexName)
			if err != nil {
				ms.logger.Error(err)
			}
			ms.logger.Debug("index '" + indexName + "' deleted.")
			// create index
			_, err = ms.client.CreateIndex(&meilisearch.IndexConfig{
				Uid:        indexName,
				PrimaryKey: "id",
			})
			if err != nil {
				ms.logger.Error(err)
			}
			ms.logger.Debug("index '" + indexName + "' created.")
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
			ms.logger.Debug("index '" + indexName + "' attributes set.")
		}
	})
	return nil
}

func (ms *MSearch) RebuildAllIndexes(dbConnUri string, dbName string) error {
	//rebuild all indexes
	ms.logger.Debug("Rebuilding all indexes...")
	dbClient, err := omgo.NewClient(context.Background(), &omgo.Config{
		Uri:      dbConnUri,
		Database: dbName,
	})
	if err != nil {
		return err
	}
	if err := dbClient.Ping(10); err != nil {
		return err
	}
	defer dbClient.Close(context.Background())
	db := dbClient.Database(dbName)

	for indexName, _ := range *ms.indexConfig.ConfigData {
		cur := db.Collection(indexName).Find(context.Background(), bson.D{}).Cursor()
		curRes := bson.M{}
		for cur.Next(curRes) {
			if curRes["is_deleted"] != nil {
				if curRes["is_deleted"].(bool) {
					continue
				}
			}
			//modify the data map
			//change _id to id
			curRes["id"] = curRes["_id"]
			delete(curRes, "_id")

			//add doc to index
			_, err := ms.client.Index(indexName).AddDocuments(&curRes)
			if err != nil {
				ms.logger.Error("Rebuild index (", indexName, ") failed: ", err)
			}
		}
	}
	ms.logger.Debug("Rebuild all indexes successfully.")
	return nil
}
