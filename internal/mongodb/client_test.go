package mongodb_test

import (
	"github.com/stretchr/testify/require"
	"go.oease.dev/goe/internal/mongodb"
	"testing"
)

func SetupDefaultConnection() *mongodb.MongoDB {
	mdb, err := mongodb.NewMongoDB("mongodb://localhost:27017/", "wshop_test")
	if err != nil {
		panic(err)
	}
	return mdb
}

func TestNewMongoDB(t *testing.T) {
	mdb, err := mongodb.NewMongoDB("mongodb://localhost:27017/", "wshop_test")
	require.Nil(t, err)
	require.NotNil(t, mdb)
}

func TestSetupWrongConnection(t *testing.T) {
	mdb, err := mongodb.NewMongoDB("wrong://localhost:27017/", "wshop_test")
	require.NotNil(t, err)
	require.Nil(t, mdb)
}

func TestPing(t *testing.T) {
	mdb := SetupDefaultConnection()
	err := mdb.Ping()
	require.Nil(t, err)
}
