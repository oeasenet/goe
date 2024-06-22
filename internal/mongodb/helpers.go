package mongodb

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.oease.dev/omgo"
)

func IsNoResult(err error) bool {
	if errors.Is(err, mongo.ErrNoDocuments) || errors.Is(err, omgo.ErrNoSuchDocuments) {
		return true
	}
	return false
}

func MustHexToObjectId(strId string) primitive.ObjectID {
	objId, err := primitive.ObjectIDFromHex(strId)
	if err != nil {
		return primitive.NilObjectID
	}
	return objId
}
