package alarmsensors

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	storage "github.com/a-castellano/AlarmSensors/storage"
)

func RetriveChildTopic(wildcardTopic string, topic string) string {
	return strings.TrimPrefix(wildcardTopic, topic)
}

func CheckSensorTriggered(ctx context.Context, sensorName string, payload string, storageInstance storage.Storage) (string, bool, error) {

	var activated bool = false
	var message string

	var sensorData map[string]interface{}

	err := json.Unmarshal([]byte(payload), &sensorData)
	if err != nil {
		return message, activated, err
	}
	// Check if sensor type is conectat one
	if _, isContactSensor := sensorData["contact"]; isContactSensor {
		sensorValue := sensorData["contact"].(bool)
		changed, _ := storageInstance.UpdateAndNotify(ctx, sensorName, sensorValue)
		if changed == true {
			if sensorValue == false {
				message = fmt.Sprintf("Contact sensor '%s' has been opened.", sensorName)
				activated = true
			} else {
				message = fmt.Sprintf("Contact sensor '%s' has been closed.", sensorName)
			}
		}
	}
	// Check is sensor type is motion
	if _, isMotionSensor := sensorData["occupancy"]; isMotionSensor {
		sensorValue := sensorData["occupancy"].(bool)
		changed, _ := storageInstance.UpdateAndNotify(ctx, sensorName, sensorValue)
		if changed == true {
			if sensorValue == true {
				message = fmt.Sprintf("Motion sensor '%s' has been triggered.", sensorName)
				activated = true
			} else {
				message = fmt.Sprintf("Motion sensor '%s' has been triggered.", sensorName)
			}
		}
	}
	return message, activated, nil
}
