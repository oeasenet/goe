package contracts

import (
	"go.oease.dev/goe/modules/mongodb"
	"go.oease.dev/omgo"
)

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
}
