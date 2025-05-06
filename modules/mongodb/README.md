# MongoDB Module

The MongoDB module provides a simple and powerful interface for working with MongoDB in the GOE framework. It implements the [`contracts.MongoDB`](https://github.com/oeasenet/goe/blob/main/contracts/mongodb.go) interface and is built on top of the official MongoDB Go driver with additional features.

## Features

- Simple API for common MongoDB operations
- Support for CRUD operations
- Query building with filters and options
- Pagination support
- Automatic document ID generation
- Automatic timestamps for creation and updates
- Meilisearch integration for full-text search
- Structured logging of database operations

## Usage

### Initialization

The MongoDB module is automatically initialized by the GOE framework if MongoDB is enabled:

```
# Enable MongoDB in your configuration
MONGODB_ENABLED=true

# Configure MongoDB connection
MONGODB_URI=mongodb://localhost:27017
MONGODB_DB=myapp
```

### Models

To work with the MongoDB module, you need to define models that implement the [`IDefaultModel`](https://github.com/oeasenet/goe/blob/main/modules/mongodb/model.go) interface. The easiest way is to embed the `DefaultModel` struct:

```go
import (
    "go.oease.dev/goe/modules/mongodb"
)

// User represents a user in the system
type User struct {
    mongodb.DefaultModel `bson:",inline"` // Embed DefaultModel to implement IDefaultModel
    Name                 string   `bson:"name" json:"name"`
    Email                string   `bson:"email" json:"email"`
    Age                  int      `bson:"age" json:"age"`
    Roles                []string `bson:"roles" json:"roles"`
}

// ColName returns the MongoDB collection name for this model
func (u *User) ColName() string {
    return "users"
}
```

The `DefaultModel` struct provides:

- `Id` field of type `primitive.ObjectID` (stored as `_id` in MongoDB)
- `CreateTime` and `LastModifyTime` fields for tracking document timestamps
- Implementation of required methods from the `IDefaultModel` interface
- Hooks for automatically setting IDs and timestamps before insert/update operations

### Basic Operations

```go
import (
    "go.mongodb.org/mongo-driver/bson"
    "go.oease.dev/goe"
)

// Get the MongoDB client
db := goe.UseDB()

// Insert a document
user := &User{
    Name:  "John Doe",
    Email: "john@example.com",
    Age:   30,
    Roles: []string{"user"},
}
result, err := db.Insert(user)
if err != nil {
    // Handle error
}
userID := user.GetId() // Get the inserted document's ID

// Find a document by ID
var foundUser User
found, err := db.FindById(&User{}, userID, &foundUser)
if err != nil {
    // Handle error
}
if !found {
    // Document not found
}

// Find documents with a filter
var users []User
err = db.Find(&User{}, bson.M{"age": bson.M{"$gt": 18}}).All(&users)
if err != nil {
    // Handle error
}

// Update a document
foundUser.Name = "John Smith"
err = db.Update(&foundUser)
if err != nil {
    // Handle error
}

// Delete a document
err = db.Delete(&foundUser)
if err != nil {
    // Handle error
}

// Check if a document exists
exists, err := db.IsExist(&User{}, bson.M{"email": "john@example.com"})
if err != nil {
    // Handle error
}

// Count documents
count, err := db.Count(&User{}, bson.M{"age": bson.M{"$gt": 18}})
if err != nil {
    // Handle error
}
```

### Pagination

The MongoDB module provides a simple way to paginate results:

```go
// Get paginated results
page := 1
pageSize := 10
var users []User
totalDocs, totalPages := db.FindPage(&User{}, bson.M{"age": bson.M{"$gt": 18}}, &users, pageSize, page)

// Access pagination information
fmt.Printf("Found %d users across %d pages\n", totalDocs, totalPages)

// With sorting options
option := mongodb.NewFindPageOption().
    SetSelector(bson.M{"name": 1, "email": 1}). // Select only name and email fields
    AddSort("age", -1).                         // Sort by age descending
    AddSort("name", 1)                          // Then by name ascending

totalDocs, totalPages = db.FindPage(&User{}, bson.M{"roles": "admin"}, &users, pageSize, page, option)
```

### Advanced Queries

```go
// Find with query builder
var users []User
err = db.Find(&User{}, bson.M{"roles": "admin"}).
    Select(bson.M{"name": 1, "email": 1}). // Select only name and email fields
    Sort(bson.M{"name": 1}).               // Sort by name ascending
    Skip(10).                              // Skip first 10 documents
    Limit(5).                              // Limit to 5 documents
    All(&users)

// Using cursor for efficient iteration
cursor := db.FindWithCursor(&User{}, bson.M{"age": bson.M{"$gt": 18}})
var user User
for cursor.Next(&user) {
    // Process each user
    fmt.Printf("User: %s, Email: %s\n", user.Name, user.Email)
}
if cursor.Err() != nil {
    // Handle error
}
cursor.Close()

// Aggregation
type AgeGroup struct {
    ID    int `bson:"_id" json:"age_group"`
    Count int `bson:"count" json:"count"`
}
var results []AgeGroup
pipeline := []bson.M{
    {"$group": bson.M{"_id": bson.M{"$floor": bson.M{"$divide": []interface{}{"$age", 10}}}, "count": bson.M{"$sum": 1}}},
    {"$sort": bson.M{"_id": 1}},
}
err = db.Aggregate(&User{}, pipeline, &results)
```

### Bulk Operations

```go
// Insert multiple documents
users := []any{
    &User{Name: "Alice", Email: "alice@example.com", Age: 25, Roles: []string{"user"}},
    &User{Name: "Bob", Email: "bob@example.com", Age: 30, Roles: []string{"admin"}},
    &User{Name: "Charlie", Email: "charlie@example.com", Age: 35, Roles: []string{"user"}},
}
result, err := db.InsertMany(&User{}, users)
if err != nil {
    // Handle error
}
fmt.Printf("Inserted %d documents\n", len(result.InsertedIDs))

// Delete multiple documents
result, err := db.DeleteMany(&User{}, bson.M{"age": bson.M{"$lt": 18}})
if err != nil {
    // Handle error
}
fmt.Printf("Deleted %d documents\n", result.DeletedCount)
```

### Meilisearch Integration

If Meilisearch is enabled, the MongoDB module can automatically sync documents to Meilisearch for full-text search:

```
# Enable Meilisearch integration in your configuration
MEILISEARCH_ENABLED=true
MEILISEARCH_DB_SYNC=true

# Configure Meilisearch connection
MEILISEARCH_ENDPOINT=http://localhost:7700
MEILISEARCH_API_KEY=your_api_key
```

When this feature is enabled:
1. Documents inserted into MongoDB are automatically indexed in Meilisearch
2. Documents updated in MongoDB are automatically updated in Meilisearch
3. Documents deleted from MongoDB are automatically removed from Meilisearch

## API Reference

### MongoDB Interface

The MongoDB module implements the [`contracts.MongoDB`](https://github.com/oeasenet/goe/blob/main/contracts/mongodb.go) interface:

```go
type MongoDB interface {
    Find(model mongodb.IDefaultModel, filter any) omgo.QueryI
    FindPage(model mongodb.IDefaultModel, filter any, res any, pageSize int64, currentPage int64, option ...*mongodb.FindPageOption) (totalDoc int64, totalPage int64)
    FindOne(model mongodb.IDefaultModel, filter any, res any) (bool, error)
    FindById(model mongodb.IDefaultModel, id string, res any) (bool, error)
    FindWithCursor(model mongodb.IDefaultModel, filter any) omgo.CursorI
    Insert(model mongodb.IDefaultModel) (*omgo.InsertOneResult, error)
    InsertMany(model mongodb.IDefaultModel, docs []any) (*omgo.InsertManyResult, error)
    Update(model mongodb.IDefaultModel) error
    Delete(model mongodb.IDefaultModel) error
    DeleteMany(model mongodb.IDefaultModel, filter any) (*omgo.DeleteResult, error)
    Aggregate(model mongodb.IDefaultModel, pipeline any, res any) error
    IsExist(model mongodb.IDefaultModel, filter any) (bool, error)
    Count(model mongodb.IDefaultModel, filter any) (int64, error)
}
```

### IDefaultModel Interface

All models must implement the [`IDefaultModel`](https://github.com/oeasenet/goe/blob/main/modules/mongodb/model.go) interface:

```go
type IDefaultModel interface {
    ColName() string                 // Returns the name of the MongoDB collection
    GetId() string                   // Returns the string representation of the ID
    GetObjectID() primitive.ObjectID // Returns the ObjectID
    PutId(id string)                 // Sets the ID using a string
    setDefaultCreateTime()           // Sets the creation time to current time
    setDefaultLastModifyTime()       // Sets the modification time to current time
    setDefaultId()                   // Sets the ID to a new ObjectID if zero
}
```

## Implementation Details

The MongoDB module provides a simplified interface for working with MongoDB while maintaining the flexibility of the underlying driver. It includes:

- Automatic document ID generation if not provided
- Automatic handling of creation and modification timestamps
- Integration with the logging module for query logging
- Optional integration with Meilisearch for full-text search
- Connection pooling and management

The implementation is based on the official MongoDB Go driver and provides additional features to make working with MongoDB easier and more productive.
