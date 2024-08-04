package mongodb

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.oease.dev/omgo"
	"go.oease.dev/omgo/options"
)

// col is a helper function that returns a MongoDB collection based on the provided model or collection name.
func (m *MongoDB) col(nameORmodel any) *omgo.Collection {
	//check if model is the default model interface
	model, ok := nameORmodel.(IDefaultModel)
	if ok && model.ColName() != "" && model.ColName() != "not_implemented" {
		return m.client.Database(m.dbName).Collection(model.ColName())
	}

	// check if model is a string
	colNameStr, ok := nameORmodel.(string)
	if ok && colNameStr != "" && colNameStr != "not_implemented" {
		return m.client.Database(m.dbName).Collection(colNameStr)
	}

	m.logger.Error("param is not a string or does not implement IDefaultModel interface")
	return nil
}

// ctx is a helper function that returns a new context with a default timeout.
func (m *MongoDB) ctx() context.Context {
	return m.newCtx()
}

// Find retrieves documents from a MongoDB collection based on the provided filter and model.
// It returns an omgo.QueryI interface that can be used to further specify the query.
// If the MongoDB instance is not initialized, it logs an error and returns nil.
// If the filter is nil, it uses an empty filter (bson.D{}) by default.
// The returned query can be used to add projections, limits, sorting, and other query modifiers.
// Example usage:
//
//	query := mongo.Find(model, bson.M{"name": "John"})
//	query.Select(bson.M{"age": 1})
//	query.Sort(bson.M{"age": -1})
//	query.Skip(5)
//	query.Limit(10)
//	cursor := query.Cursor()
//	for cursor.Next(context.Background()) {
//	  var result *MyModel
//	  cursor.Decode(&result)
//	  // process the result
//	}
func (m *MongoDB) Find(model IDefaultModel, filter any) omgo.QueryI {
	if !m.initialized {
		m.logger.Error("Must initialize MongoDB first, by calling NewMongodb() method")
		return nil
	}

	if filter == nil {
		filter = bson.D{}
	}

	return m.col(model).Find(m.ctx(), filter)
}

// FindPage is a method that finds a page of documents in a MongoDB collection based on the provided filter.
// It returns the total number of documents and the total number of pages.
func (m *MongoDB) FindPage(model IDefaultModel, filter any, res any, pageSize int64, currentPage int64, option ...*FindPageOption) (totalDoc int64, totalPage int64) {
	if !m.initialized {
		m.logger.Error("Must initialize MongoDB first, by calling NewMongodb() method")
		return 0, 0
	}

	var opt *FindPageOption
	if len(option) > 0 && option[0] != nil {
		opt = option[0]
	} else {
		opt = nil
	}

	if filter == nil {
		filter = bson.D{}
	}

	countDoc, err := m.col(model).Find(m.ctx(), filter).Count()
	if IsNoResult(err) {
		res = nil
		return 0, 0
	}
	if err != nil {
		res = nil
		m.logger.Error(err)
		return 0, 0
	}

	//calculate the offset of how many documents to skip
	offset := (currentPage - 1) * pageSize
	//calculate how many documents to show in the current page
	var limit int64
	if countDoc-offset < pageSize {
		limit = countDoc - offset
	} else {
		limit = pageSize
	}

	//calculate the total page
	if countDoc%pageSize == 0 {
		totalPage = countDoc / pageSize
	} else {
		totalPage = countDoc/pageSize + 1
	}

	//find the documents
	query := m.col(model).Find(m.ctx(), filter)
	if opt != nil {
		query.Select(opt.selector)
		query.Sort(opt.fields...)
	}
	err = query.Limit(limit).Skip(offset).All(res)
	if IsNoResult(err) {
		res = nil
		return 0, 0
	}
	if err != nil {
		res = nil
		m.logger.Error(err)
		return 0, 0
	}
	releaseFindPageOption(opt)
	return countDoc, totalPage
}

// FindOne is a method that finds a single document in a MongoDB collection based on the provided filter.
func (m *MongoDB) FindOne(model IDefaultModel, filter any, res any) (bool, error) {
	if !m.initialized {
		return false, errors.New("must initialize MongoDB first, by calling NewMongodb() method")
	}

	if filter == nil {
		filter = bson.D{}
	}
	err := m.col(model).Find(m.ctx(), filter).One(res)
	if IsNoResult(err) {
		res = nil
		return false, nil
	}
	if err != nil {
		res = nil
		return false, err
	}
	return true, nil
}

// FindById is a method that finds a single document in a MongoDB collection based on the provided id.
func (m *MongoDB) FindById(model IDefaultModel, id string, res any) (bool, error) {
	if !m.initialized {
		return false, errors.New("must initialize MongoDB first, by calling NewMongodb() method")
	}

	err := m.col(model).Find(m.ctx(), bson.M{"_id": MustHexToObjectId(id)}).One(res)
	if IsNoResult(err) {
		res = nil
		return false, nil
	}
	if err != nil {
		res = nil
		return false, err
	}

	return true, nil
}

// FindWithCursor is a method that returns an omgo.CursorI interface based on the provided model and filter.
// If the MongoDB instance is not initialized, it logs an error and returns nil.
// If the filter is nil, it uses an empty filter (bson.D{}) by default.
// It returns the cursor object.
func (m *MongoDB) FindWithCursor(model IDefaultModel, filter any) omgo.CursorI {
	if !m.initialized {
		m.logger.Error("Must initialize MongoDB first, by calling NewMongodb() method")
		return nil
	}

	if filter == nil {
		filter = bson.D{}
	}

	return m.col(model).Find(m.ctx(), filter).Cursor()
}

// Insert is a method that inserts a single document into a MongoDB collection.
func (m *MongoDB) Insert(model IDefaultModel) (*omgo.InsertOneResult, error) {
	if !m.initialized {
		return nil, errors.New("must initialize MongoDB first, by calling NewMongodb() method")
	}

	return m.col(model).InsertOne(m.ctx(), model, options.InsertOneOptions{InsertHook: model})
}

// InsertMany inserts multiple documents into a MongoDB collection based on the provided model and slice of documents.
// It returns the result of the insert operation and an error if any.
// If the MongoDB instance is not initialized, it logs an error and returns nil.
// Example usage:
//
//	model := MyModel{}
//	docs := []MyModel{
//		{Name: "Jane", Age: 25},
//		{Name: "Doe", Age: 35},
//	}
//	result, err := mongo.InsertMany(&model, docs)
//	if err != nil {
//		// handle error
//	}
//	fmt.Println(result.InsertedIDs)
//
// Parameter model: The model or collection name.
// Parameter docs: The slice of documents to be inserted.
// Returns: *omgo.InsertManyResult, error
func (m *MongoDB) InsertMany(model IDefaultModel, docs []any) (*omgo.InsertManyResult, error) {
	if !m.initialized {
		m.logger.Error("Must initialize MongoDB first, by calling NewMongodb() method")
		return nil, errors.New("must initialize MongoDB first, by calling NewMongodb() method")
	}

	return m.col(model).InsertMany(m.ctx(), docs, options.InsertManyOptions{InsertHook: model})
}

// Update is a method that updates a single document in a MongoDB collection.
func (m *MongoDB) Update(model IDefaultModel) error {
	if !m.initialized {
		return errors.New("must initialize MongoDB first, by calling NewMongodb() method")
	}
	// create a filter
	f := bson.M{}

	// check if model has an ID or has the document been found
	if model.GetId() == "" || model.GetObjectID() == primitive.NilObjectID || model.GetObjectID().IsZero() {
		return errors.New("model does not have an ID, please provide an ID or find the document first")
	}

	// add the ID to the filter
	f["_id"] = model.GetObjectID()

	return m.col(model).UpdateOne(m.ctx(), f, bson.M{"$set": model}, options.UpdateOptions{UpdateHook: model})
}

// Delete is a method that deletes a single document from a MongoDB collection.
func (m *MongoDB) Delete(model IDefaultModel) error {
	if !m.initialized {
		return errors.New("must initialize MongoDB first, by calling NewMongodb() method")
	}

	// check if model has an ID or has the document been found
	if model.GetId() == "" || model.GetObjectID() == primitive.NilObjectID || model.GetObjectID().IsZero() {
		return errors.New("model does not have an ID, please provide an ID or find the document first")
	}

	return m.col(model).RemoveId(m.ctx(), model.GetObjectID())
}

// Aggregate is a method that performs an aggregation pipeline operation on a MongoDB collection.
func (m *MongoDB) Aggregate(model IDefaultModel, pipeline any, res any) error {
	if !m.initialized {
		return errors.New("must initialize MongoDB first, by calling NewMongodb() method")
	}

	return m.col(model).Aggregate(m.ctx(), pipeline).All(res)
}
