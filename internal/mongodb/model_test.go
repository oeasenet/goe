package mongodb

import (
	"context"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
	"time"
)

func TestDefaultModel_ColName(t *testing.T) {
	model := &DefaultModel{}
	require.Equal(t, "not_implemented", model.ColName())
}

func TestDefaultModel_GetId(t *testing.T) {
	id := primitive.NewObjectID()
	model := &DefaultModel{Id: id}
	require.Equal(t, id.Hex(), model.GetId())
}

func TestDefaultModel_PutId(t *testing.T) {
	id := primitive.NewObjectID().Hex()
	model := &DefaultModel{}
	model.PutId(id)
	require.Equal(t, id, model.GetId())
}

func TestDefaultModel_setDefaultCreateTime(t *testing.T) {
	model := &DefaultModel{}
	model.setDefaultCreateTime()
	require.LessOrEqual(t, time.Now().UnixMilli()-model.CreateTime, int64(1000))
}

func TestDefaultModel_setDefaultLastModifyTime(t *testing.T) {
	model := &DefaultModel{}
	model.setDefaultLastModifyTime()
	require.LessOrEqual(t, time.Now().UnixMilli()-model.LastModifyTime, int64(1000))
}

func TestDefaultModel_setDefaultId(t *testing.T) {
	model := &DefaultModel{}
	model.setDefaultId()
	require.False(t, model.Id.IsZero())
}

func TestDefaultModel_BeforeInsert(t *testing.T) {
	model := &DefaultModel{}
	err := model.BeforeInsert(context.Background())
	require.Nil(t, err)
	require.False(t, model.Id.IsZero())
	require.LessOrEqual(t, time.Now().UnixMilli()-model.CreateTime, int64(1000))
	require.LessOrEqual(t, time.Now().UnixMilli()-model.LastModifyTime, int64(1000))
}

func TestDefaultModel_BeforeUpdate(t *testing.T) {
	model := &DefaultModel{}
	err := model.BeforeUpdate(context.Background())
	require.Nil(t, err)
	require.LessOrEqual(t, time.Now().UnixMilli()-model.LastModifyTime, int64(1000))
}

func TestDefaultModel_BeforeUpsert(t *testing.T) {
	model := &DefaultModel{}
	err := model.BeforeUpsert(context.Background())
	require.Nil(t, err)
	require.False(t, model.Id.IsZero())
	require.LessOrEqual(t, time.Now().UnixMilli()-model.CreateTime, int64(1000))
	require.LessOrEqual(t, time.Now().UnixMilli()-model.LastModifyTime, int64(1000))
}

func TestDefaultModel_GetObjectID(t *testing.T) {
	id := primitive.NewObjectID()
	model := &DefaultModel{Id: id}
	require.Equal(t, id, model.GetObjectID())
}
