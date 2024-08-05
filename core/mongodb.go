package core

import (
	"errors"
	"go.oease.dev/goe/modules/mongodb"
	"go.oease.dev/goe/modules/msearch"
	"go.oease.dev/omgo"
)

type GoeMongoDB struct {
	goeConfig       *GoeConfig
	msearchInstance *msearch.MSearch
	mongodbInstance *mongodb.MongoDB
}

func NewGoeMongoDB(appConfig *GoeConfig, logger mongodb.Logger) (*GoeMongoDB, error) {
	if appConfig.MongoDB.DB == "" {
		return nil, errors.New("database name is required")
	}
	if appConfig.MongoDB.URI == "" {
		return nil, errors.New("connection uri is required")
	}
	mdb, err := mongodb.NewMongoDB(appConfig.MongoDB.URI, appConfig.MongoDB.DB, logger)
	if err != nil {
		return nil, err
	}
	return &GoeMongoDB{
		goeConfig:       appConfig,
		mongodbInstance: mdb,
		msearchInstance: nil,
	}, nil
}

func (g *GoeMongoDB) SetMeilisearch(meilisearch *msearch.MSearch) error {
	if !g.goeConfig.Features.MeilisearchEnabled {
		return errors.New("meilisearch is not enabled")
	}
	if !g.goeConfig.Features.SearchDBSyncEnabled {
		return errors.New("meilisearch db sync is not enabled")
	}
	g.msearchInstance = meilisearch
	return nil
}

func (g *GoeMongoDB) Find(model mongodb.IDefaultModel, filter any) omgo.QueryI {
	return g.mongodbInstance.Find(model, filter)
}

func (g *GoeMongoDB) FindPage(model mongodb.IDefaultModel, filter any, res any, pageSize int64, currentPage int64, option ...*mongodb.FindPageOption) (totalDoc int64, totalPage int64) {
	return g.mongodbInstance.FindPage(model, filter, res, pageSize, currentPage, option...)
}

func (g *GoeMongoDB) FindOne(model mongodb.IDefaultModel, filter any, res any) (bool, error) {
	return g.mongodbInstance.FindOne(model, filter, res)
}

func (g *GoeMongoDB) FindById(model mongodb.IDefaultModel, id string, res any) (bool, error) {
	return g.mongodbInstance.FindById(model, id, res)
}

func (g *GoeMongoDB) FindWithCursor(model mongodb.IDefaultModel, filter any) omgo.CursorI {
	return g.mongodbInstance.FindWithCursor(model, filter)
}

func (g *GoeMongoDB) Insert(model mongodb.IDefaultModel) (*omgo.InsertOneResult, error) {
	ior, err := g.mongodbInstance.Insert(model)
	if g.goeConfig.Features.MeilisearchEnabled && g.goeConfig.Features.SearchDBSyncEnabled {
		if g.msearchInstance != nil {
			err := g.msearchInstance.AddDoc(model.ColName(), model)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New("meilisearch instance is not set")
		}
	}
	return ior, err
}

func (g *GoeMongoDB) InsertMany(model mongodb.IDefaultModel, docs []any) (*omgo.InsertManyResult, error) {
	imr, err := g.mongodbInstance.InsertMany(model, docs)
	if g.goeConfig.Features.MeilisearchEnabled && g.goeConfig.Features.SearchDBSyncEnabled {
		if g.msearchInstance != nil {
			for _, doc := range docs {
				err := g.msearchInstance.AddDoc(model.ColName(), doc)
				if err != nil {
					return nil, err
				}
			}
		} else {
			return nil, errors.New("meilisearch instance is not set")
		}
	}
	return imr, err
}

func (g *GoeMongoDB) Update(model mongodb.IDefaultModel) error {
	e := g.mongodbInstance.Update(model)
	if g.goeConfig.Features.MeilisearchEnabled && g.goeConfig.Features.SearchDBSyncEnabled {
		if g.msearchInstance != nil {
			err := g.msearchInstance.UpdateDoc(model.ColName(), model)
			if err != nil {
				return err
			}
		} else {
			return errors.New("meilisearch instance is not set")
		}
	}
	return e
}

func (g *GoeMongoDB) Delete(model mongodb.IDefaultModel) error {
	e := g.mongodbInstance.Delete(model)
	if g.goeConfig.Features.MeilisearchEnabled && g.goeConfig.Features.SearchDBSyncEnabled {
		if g.msearchInstance != nil {
			err := g.msearchInstance.DelDoc(model.ColName(), model.GetId())
			if err != nil {
				return err
			}
		} else {
			return errors.New("meilisearch instance is not set")
		}
	}
	return e
}

func (g *GoeMongoDB) Aggregate(model mongodb.IDefaultModel, pipeline any, res any) error {
	return g.mongodbInstance.Aggregate(model, pipeline, res)
}
