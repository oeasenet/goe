package contracts

import "github.com/meilisearch/meilisearch-go"

type Meilisearch interface {
	ApplyIndexConfigs(configData []byte) error
	WaitForTaskSuccess(taskUID int64) error
	AddDoc(indexName string, docPtr any) error
	DelDoc(indexName string, docId string) error
	UpdateDoc(indexName string, docPtr any) error
	GetDoc(indexName string, docId string, bindResult any) (bool, error)
	DeleteAllDocuments(indexName string) error
	Search(indexName string, query string, options *meilisearch.SearchRequest) *meilisearch.SearchResponse
}
