package mongodb

import (
	"context"
	"errors"
)

type updater struct {
	collectionModel IDefaultModel
	ctx             context.Context
	m               *Mongodb
	hasResult       bool
	err             error
}

func (m *Mongodb) UseUpdater(model IDefaultModel) *updater {
	return &updater{
		collectionModel: model,
		ctx:             m.ctx(),
		hasResult:       false,
	}
}

func (u *updater) Find() (*updater, bool) {
	if u.collectionModel == nil {
		u.err = errors.New("must provide a valid model to updater")
		return u, false
	}

	hasResult, err := u.m.FindById(u.collectionModel, u.collectionModel.GetId(), u.collectionModel)
	if !hasResult {
		u.err = nil
		u.hasResult = false
		return u, false
	}
	if err != nil {
		u.err = err
		return u, false
	}
	u.hasResult = true
	return u, true
}

func (u *updater) DoUpdate() error {
	if !u.hasResult {
		return errors.New("document does not exist")
	}
	if u.collectionModel == nil {
		return errors.New("must provide a valid model to updater")
	}
	if u.err != nil {
		return u.err
	}

	return u.m.Update(u.collectionModel)
}

func (u *updater) Err() error {
	return u.err
}
