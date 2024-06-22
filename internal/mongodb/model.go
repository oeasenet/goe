package mongodb

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// IDefaultModel is an interface that defines the methods required for a default model.
// The interface includes methods to retrieve and manipulate the ID, as well as the creation
// and modification times of the model. Additionally, it provides a method to obtain the
// MongoDB collection name associated with the model.
type IDefaultModel interface {
	ColName() string                 // Returns the name of the MongoDB collection associated with the model.
	GetId() string                   // Returns the string representation of the model's ID.
	GetObjectID() primitive.ObjectID // Returns the ObjectID of the model.
	PutId(id string)                 // Sets the model's ID using a string.
	setDefaultCreateTime()           // Sets the model's creation time to the current time.
	setDefaultLastModifyTime()       // Sets the model's modification time to the current time.
	setDefaultId()                   // Sets the model's ID to a new ObjectID if it is currently zero.
}

// DefaultModel is a struct that represents a default model for MongoDB documents.
// It includes fields for the document's ID, creation time, and modification time.
type DefaultModel struct {
	Id             primitive.ObjectID `bson:"_id" json:"id"`                            // The document's ID.
	CreateTime     int64              `bson:"create_time" json:"create_time"`           // The document's creation time.
	LastModifyTime int64              `bson:"last_modify_time" json:"last_modify_time"` // The document's modification time.
}

// ColName returns the name of the collection to which the DefaultModel belongs.
// Currently, this method returns a placeholder string.
func (m *DefaultModel) ColName() string {
	return "not_implemented"
}

// GetId returns the hexadecimal string representation of the Id field of the DefaultModel.
// It uses the Hex() method from the Id field to convert the Id to its string representation.
func (m *DefaultModel) GetId() string {
	return m.Id.Hex()
}

// PutId sets the Id field of the DefaultModel using a hexadecimal string.
func (m *DefaultModel) PutId(id string) {
	hex, _ := primitive.ObjectIDFromHex(id)
	m.Id = hex
}

// setDefaultCreateTime sets the CreateTime field of the DefaultModel to the current time in milliseconds.
func (m *DefaultModel) setDefaultCreateTime() {
	m.CreateTime = time.Now().UnixMilli()
}

// setDefaultLastModifyTime sets the LastModifyTime field of the DefaultModel to the current time in milliseconds.
func (m *DefaultModel) setDefaultLastModifyTime() {
	m.LastModifyTime = time.Now().UnixMilli()
}

// setDefaultId sets the Id field of the DefaultModel to a new ObjectID if it is currently zero.
func (m *DefaultModel) setDefaultId() {
	if m.Id.IsZero() {
		m.Id = primitive.NewObjectID()
	}
}

// BeforeInsert is a method that is called before a document is inserted into the database.
// It sets the document's ID and timestamps to their default values.
func (m *DefaultModel) BeforeInsert(ctx context.Context) error {
	m.setDefaultId()
	m.setDefaultCreateTime()
	m.setDefaultLastModifyTime()
	return nil
}

// BeforeUpdate is a method that is called before a document is updated in the database.
// It sets the document's modification time to the current time.
func (m *DefaultModel) BeforeUpdate(ctx context.Context) error {
	m.setDefaultLastModifyTime()
	return nil
}

//// BeforeUpsert is a method that is called before a document is upserted into the database.
//// It sets the document's ID and timestamps to their default values.
//func (m *DefaultModel) BeforeUpsert(ctx context.Context) error {
//	m.setDefaultId()
//	m.setDefaultCreateTime()
//	m.setDefaultLastModifyTime()
//	return nil
//}

// GetObjectID returns the ObjectID of the DefaultModel.
func (m *DefaultModel) GetObjectID() primitive.ObjectID {
	return m.Id
}
