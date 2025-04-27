# MSearch Module

The MSearch module provides integration with [Meilisearch](https://www.meilisearch.com/), a powerful, fast, and open-source search engine. It implements the `contracts.Meilisearch` interface and offers full-text search capabilities for your application.

## Features

- Full-text search with Meilisearch
- Index management
- Document indexing and searching
- Automatic synchronization with MongoDB
- Configurable search settings
- Structured logging of search operations

## Usage

### Initialization

The MSearch module is automatically initialized by the GOE framework if Meilisearch is enabled:

```
# Enable Meilisearch in your configuration
MEILISEARCH_ENABLED=true

# Configure Meilisearch connection
MEILISEARCH_ENDPOINT=http://localhost:7700
MEILISEARCH_API_KEY=your_api_key

# Enable automatic synchronization with MongoDB (optional)
MEILISEARCH_DB_SYNC=true
```

### Basic Operations

```go
// Get the Meilisearch client
search := goe.UseSearch()

// Add a document to an index
document := map[string]interface{}{
    "id": "1",
    "title": "Getting Started with GOE",
    "content": "This is a guide to getting started with the GOE framework.",
    "tags": []string{"go", "framework", "tutorial"},
}
err := search.Index("articles").AddDocuments([]map[string]interface{}{document})

// Search for documents
results, err := search.Index("articles").Search("framework", &meilisearch.SearchRequest{
    Limit: 10,
})

// Delete a document from an index
err := search.Index("articles").DeleteDocument("1")
```

### Configuring Indexes

You can configure indexes using a JSON configuration file:

```json
{
  "indexes": [
    {
      "uid": "articles",
      "primaryKey": "id",
      "searchableAttributes": ["title", "content", "tags"],
      "displayedAttributes": ["id", "title", "content", "tags", "created_at"],
      "filterableAttributes": ["tags"],
      "sortableAttributes": ["created_at"]
    },
    {
      "uid": "users",
      "primaryKey": "id",
      "searchableAttributes": ["name", "email", "bio"],
      "displayedAttributes": ["id", "name", "email", "bio", "created_at"],
      "filterableAttributes": ["role"],
      "sortableAttributes": ["created_at"]
    }
  ]
}
```

Apply the configuration:

```go
// Load the configuration from a file
configData, err := os.ReadFile("configs/msearch.json")
if err != nil {
    log.Fatal(err)
}

// Apply the configuration
err = search.ApplyIndexConfigs(configData)
if err != nil {
    log.Fatal(err)
}
```

### MongoDB Integration

When `MEILISEARCH_DB_SYNC` is enabled, the module automatically synchronizes MongoDB documents with Meilisearch:

1. When a document is inserted into MongoDB, it's also indexed in Meilisearch
2. When a document is updated in MongoDB, it's also updated in Meilisearch
3. When a document is deleted from MongoDB, it's also removed from Meilisearch

To enable this feature for a specific collection, you need to configure it in your model:

```go
type Article struct {
    ID        string    `bson:"_id" json:"id" search:"true"`
    Title     string    `bson:"title" json:"title" search:"true"`
    Content   string    `bson:"content" json:"content" search:"true"`
    Tags      []string  `bson:"tags" json:"tags" search:"true"`
    CreatedAt time.Time `bson:"created_at" json:"created_at"`
    UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// The `search:"true"` tag indicates that this field should be indexed in Meilisearch
```

## Implementation Details

The MSearch module provides a simplified interface for working with Meilisearch while maintaining access to the underlying functionality. It includes:

- Integration with the logging module for search operation logging
- Automatic synchronization with MongoDB (optional)
- Support for complex search queries
- Index configuration management

The module is designed to make it easy to add powerful full-text search capabilities to your application.