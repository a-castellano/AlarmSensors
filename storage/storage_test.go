package storage

import (
	"context"
	"testing"

	redismock "github.com/go-redis/redismock/v8"
)

func TestReadEmptySet(t *testing.T) {
	db, mock := redismock.NewClientMock()

	var key string = "ab123"
	mock.ExpectHGetAll(key).RedisNil()

	storageInstance := Storage{db}
	var ctx = context.TODO()

	changed, err := storageInstance.UpdateAndNotify(ctx, key, true)
	if err != nil {
		t.Error("TestReadEmptySet, should not fail, error was ", err.Error())
	}
	if changed != true {
		t.Error("TestReadEmptySet, changed should be true as no previous record was stored.")
	}

}

func TestNoChanged(t *testing.T) {
	db, mock := redismock.NewClientMock()

	expectedValues := make(map[string]string)
	expectedValues["name"] = "ab123"
	expectedValues["lastupdated"] = "123"
	expectedValues["triggered"] = "false"

	var key string = "ab123"
	mock.ExpectHGetAll(key).SetVal(expectedValues)

	storageInstance := Storage{db}
	var ctx = context.TODO()

	changed, err := storageInstance.UpdateAndNotify(ctx, key, false)
	if err != nil {
		t.Error("TestNoChanged, should not fail, error was ", err.Error())
	}
	if changed != false {
		t.Error("TestNoChanged, changed should be false as previous record has not changed.")
	}

}

func TestChanged(t *testing.T) {
	db, mock := redismock.NewClientMock()

	expectedValues := make(map[string]string)
	expectedValues["name"] = "ab123"
	expectedValues["lastupdated"] = "123"
	expectedValues["triggered"] = "false"

	var key string = "ab123"
	mock.ExpectHGetAll(key).SetVal(expectedValues)

	storageInstance := Storage{db}
	var ctx = context.TODO()

	changed, err := storageInstance.UpdateAndNotify(ctx, key, true)
	if err != nil {
		t.Error("TestNoChanged, should not fail, error was ", err.Error())
	}
	if changed != true {
		t.Error("TestNoChanged, changed should be true as previous record was changed.")
	}

}
