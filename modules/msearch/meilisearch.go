package msearch

import (
	"context"
	"errors"
	"github.com/meilisearch/meilisearch-go"
	"reflect"
	"time"
)

// WaitForTaskSuccess is a method that waits for a task with a specified taskUID to complete.
// It will return an error either if there's an error occurring while getting the task detail
// from the client or when the context deadline is exceeded (after 20 seconds).
//
// This function starts by polling for the status of the task every 50 milliseconds,
// and every time the task status is retrieved, the wait time for the next check is doubled,
// up to a maximum delay of 1 second. This is done to limit the number of calls
// to the GetTask function, especially in cases where the task takes a significant amount
// of time to complete.
func (ms *MSearch) WaitForTaskSuccess(taskUID int64) error {
	// Create a context with a deadline of 30 seconds.
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*30)
	// Ensure the cancel function is called once this function completes.
	defer cancelFunc()

	// Set the initial delay between status checks to 100 milliseconds.
	delay := time.Millisecond * 100
	// Set a maximum delay between status checks to 2 second.
	maxDelay := 1 * time.Second

	for {
		// Check the status of the task.
		task, err := ms.client.GetTask(taskUID)
		if err != nil {
			// Return any error that occurred while getting the task.
			return err
		}

		if task.Status == meilisearch.TaskStatusSucceeded {
			// If the task has succeeded, return nil to indicate success.
			return nil
		}

		select {
		case <-ctx.Done():
			// If the context deadline has been exceeded, return the error from the context.
			return ctx.Err()
		case <-time.After(delay):
			// Wait for the current delay period
			// then proceed to the next iteration to check the status of the task again.

			// Double the delay for the next iteration.
			// This reduces the number of calls to GetTask when the task takes a long time to complete.
			if delay < maxDelay {
				delay *= 2
				// but don't let it grow more than maxDelay.
				if delay > maxDelay {
					delay = maxDelay
				}
			}
		}
	}
}

// AddDoc is a method that adds a document to the index specified by indexName.
// The docPtr parameter must be a pointer, otherwise an error will be returned.
// The method first checks if the docPtr is a pointer using reflection.
// If the check fails, an error is returned.
// Otherwise, the method calls the AddDocuments function on the corresponding index
// through the Meilisearch client. If there's an error during the AddDocuments call,
// the error is returned.
// Finally, the method calls WaitForTaskSuccess to wait for the task to complete,
// using the TaskUID from the response of the AddDocuments call.
// If the wait is successful, the method returns nil.
//
// Note: This method depends on the MSearch.WaitForTaskSuccess method for waiting
// for the task to complete. Please refer to the documentation of that method for more details.
func (ms *MSearch) AddDoc(indexName string, docPtr any) error {
	if reflect.ValueOf(docPtr).Kind() != reflect.Ptr {
		return errors.New("docPtr must be a pointer")
	}
	resp, err := ms.client.Index(indexName).AddDocuments(docPtr)
	if err != nil {
		return err
	}
	return ms.WaitForTaskSuccess(resp.TaskUID)
}

// DelDoc is a method that deletes a document with the specified docId from the index with the specified indexName.
// It makes a call to the MeiliSearch Index DeleteDocument() method using the client.
// It returns an error if an error occurs while deleting the document, otherwise it waits for the delete task to complete
// using the WaitForTaskSuccess() method and returns any error that occurs during waiting.
// The WaitForTaskSuccess() method waits for a task with a specified taskUID to complete.
// It polls for the status of the task every 50 milliseconds, increasing the wait time for the next check by doubling it,
// up to a maximum delay of 1 second. This is done to limit the number of calls to the GetTask function when the task takes a significant amount of time to complete.
func (ms *MSearch) DelDoc(indexName string, docId string) error {
	resp, err := ms.client.Index(indexName).DeleteDocument(docId)
	if err != nil {
		return err
	}
	return ms.WaitForTaskSuccess(resp.TaskUID)
}

// UpdateDoc is a method that updates a document in the specified index with the given data.
// It first checks if the docPtr is a pointer and returns an error if it is not.
// It then calls the UpdateDocuments function of the client's Index with the docPtr.
// If there's an error during the update process, it returns the error.
// Finally, it calls WaitForTaskSuccess with the TaskUID from the response,
// to wait for the update task to complete.
func (ms *MSearch) UpdateDoc(indexName string, docPtr any) error {
	if reflect.ValueOf(docPtr).Kind() != reflect.Ptr {
		return errors.New("docPtr must be a pointer")
	}
	resp, err := ms.client.Index(indexName).UpdateDocuments(docPtr)
	if err != nil {
		return err
	}
	return ms.WaitForTaskSuccess(resp.TaskUID)
}

func (ms *MSearch) GetDoc(indexName string, docId string, bindResult any) (bool, error) {
	if reflect.ValueOf(bindResult).Kind() != reflect.Ptr {
		return false, errors.New("bindResult must be a pointer")
	}
	err := ms.client.Index(indexName).GetDocument(docId, nil, bindResult)
	if err != nil {
		if err.(*meilisearch.Error).StatusCode == 404 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Search is a method that performs a search query on the specified index with the given query and options.
// It returns a *meilisearch.SearchResponse containing the search results.
// If there's an error occurring while performing the search query, it will log the error and return nil.
func (ms *MSearch) Search(indexName string, query string, options *meilisearch.SearchRequest) *meilisearch.SearchResponse {
	resp, err := ms.client.Index(indexName).Search(query, options)
	if err != nil {
		ms.logger.Error(err)
		return nil
	}
	return resp
}
