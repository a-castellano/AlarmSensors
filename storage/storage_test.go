package storage

import (
	"context"
	"testing"

	redismock "github.com/go-redis/redismock/v8"
)

func TestNewsReadEmptySet(t *testing.T) {
	db, mock := redismock.NewClientMock()

	var key string = "ab123"
	mock.ExpectHGetAll(key).RedisNil()

	storageInstance := Storage{db}
	var ctx = context.TODO()

	changed, err := storageInstance.UpdateAndNotify(ctx, key, true)
	if err != nil {
		t.Error("TestNewsReadEmptySet, should not fail, error was ", err.Error())
	}
	if changed != true {
		t.Error("TestNewsReadEmptySet, changed should be true as no previos record was stored.")
	}

}
