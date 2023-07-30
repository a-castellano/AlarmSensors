package storage

import (
	"context"
	"time"

	goredis "github.com/go-redis/redis/v8"
)

type SensorStatus struct {
	Name        string `redis:"name"`
	LastUpdated int64  `redis:"lastupdated"`
	Triggered   bool   `redis:"triggered"`
}

type Storage struct {
	RedisClient *goredis.Client
}

func (storage Storage) UpdateAndNotify(ctx context.Context, sensorName string, sensorValue bool) (bool, error) {
	// Look for sensor stored info
	var changed bool = false
	var sensorStatus SensorStatus
	now := time.Now()
	storedSensorInfoError := storage.RedisClient.HGetAll(ctx, sensorName).Scan(&sensorStatus)
	if storedSensorInfoError == goredis.Nil {
		//Sensor info has not been stored yet
		sensorStatus.Name = sensorName
		sensorStatus.LastUpdated = now.Unix()
		sensorStatus.Triggered = sensorValue
		// Update Redis
		storage.RedisClient.HSet(ctx, sensorName, "name", sensorStatus.Name)
		storage.RedisClient.HSet(ctx, sensorName, "lastupdated", sensorStatus.LastUpdated)
		storage.RedisClient.HSet(ctx, sensorName, "triggered", sensorStatus.Triggered)
		changed = true
	} else {
		if storedSensorInfoError != nil {
			return changed, storedSensorInfoError
		}
		// Check if Triggered value differs
		if sensorValue != sensorStatus.Triggered {
			changed = true
			storage.RedisClient.HSet(ctx, sensorName, "triggered", sensorStatus.Triggered)
		}
		sensorStatus.LastUpdated = now.Unix()
		storage.RedisClient.HSet(ctx, sensorName, "lastupdated", sensorStatus.LastUpdated)
	}
	return changed, nil
}
