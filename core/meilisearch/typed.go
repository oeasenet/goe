package meilisearch

import (
	"github.com/meilisearch/meilisearch-go"
)

// Repository provides a generic repository pattern for MeiliSearch index
type Repository[T any] struct {
	index meilisearch.IndexManager
}

// NewRepository creates a new typed repository for an index
func NewRepository[T any](sm meilisearch.ServiceManager, indexName string) *Repository[T] {
	return &Repository[T]{
		index: sm.Index(indexName),
	}
}

// Index returns the underlying meilisearch index
func (r *Repository[T]) Index() meilisearch.IndexManager {
	return r.index
}

// FindByID finds a single document by id
func (r *Repository[T]) FindByID(id string, filter *meilisearch.DocumentQuery) (*T, error) {
	var result T
	err := r.index.GetDocument(id, filter, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// AddDocuments add document(s) to index
func (r *Repository[T]) AddDocuments(documentsPtr interface{}) error {
	_, err := r.index.AddDocuments(documentsPtr)
	return err
}

// UpdateDocuments update document(s) to index
func (r *Repository[T]) UpdateDocuments(documentsPtr interface{}) error {
	_, err := r.index.UpdateDocuments(documentsPtr)
	return err
}

// DeleteDocument delete single document by id
func (r *Repository[T]) DeleteDocument(id string) error {
	_, err := r.index.DeleteDocument(id)
	return err
}

// Search documents in index
func (r *Repository[T]) Search(query string, options *meilisearch.SearchRequest) (*meilisearch.SearchResponse, error) {
	response, err := r.index.Search(query, options)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// SearchPage simple search documents in index with pagination
func (r *Repository[T]) SearchPage(query string, options *meilisearch.SearchRequest) (*meilisearch.SearchResponse, error) {
	response, err := r.index.Search(query, options)
	if err != nil {
		return nil, err
	}
	return response, nil
}
