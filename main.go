package main

import (
	"fmt"
	"log/syslog"
	"net/http"
	"time"

	alarmsensors "github.com/a-castellano/AlarmSensors/alarmsensors"
	config "github.com/a-castellano/AlarmSensors/config_reader"
	apiwatcher "github.com/a-castellano/AlarmStatusWatcher/apiwatcher"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

func sub(client mqtt.Client, topicFromConfig string, syslog *syslog.Writer) {
	topic := fmt.Sprintf("%s+", topicFromConfig)
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	syslog.Info(fmt.Sprintf("Subscribed to topic: %s", topic))
}

func handleMessage(serviceConfig config.Config, syslog *syslog.Writer, watcher apiwatcher.APIWatcher, alarmManagerRequester apiwatcher.Requester, topic string, message string) {

	fmt.Printf("RECEIVED TOPIC: %s MESSAGE: %s\n", topic, message)
	candidateSensor := alarmsensors.RetriveChildTopic(topic, serviceConfig.Mqtt.WildcardTopic)

	if _, sensorIsManaged := serviceConfig.Sensors[candidateSensor]; sensorIsManaged {
		statusMessage, sensorActivated, checkSensorErr := alarmsensors.CheckSensorTriggered(candidateSensor, message)
		if checkSensorErr != nil {
			errorString := fmt.Sprintf("%v", checkSensorErr.Error())
			syslog.Err(errorString)
		} else {
			fmt.Println(statusMessage)
			syslog.Info(statusMessage)
			// Check alarm status
			fmt.Println(sensorActivated)
			if sensorActivated == true {
				apiInfo, apiInfoErr := watcher.ShowInfo(alarmManagerRequester)
				currentAlarmMode := apiInfo.DevicesInfo[serviceConfig.AlarmManager.DeviceId].Mode
				// Check if sensor triggers alarm
				fmt.Println(serviceConfig.Sensors[candidateSensor].SensorTriggers)
				if _, triggerAlarm := serviceConfig.Sensors[candidateSensor].SensorTriggers[currentAlarmMode]; triggerAlarm {
					logMessage := fmt.Sprintf("%s sensor has been triggered and alarm status is %s, triggering alarm.", candidateSensor, currentAlarmMode)
					syslog.Info(logMessage)
					//					jsonString := fmt.Sprintf("{\"mode\":\"SOS\"}")
					//					var jsonStr = []byte(jsonString)
					//					apiURL := fmt.Sprintf("http://%s:%d/devices/status/%s", serviceConfig.AlarmManager.Host, serviceConfig.AlarmManager.Port, serviceConfig.AlarmManager.DeviceId)
					//					req, _ := http.NewRequest("PUT", apiURL, bytes.NewBuffer(jsonStr))
					//					req.Header.Set("Content-Type", "application/json")
					//					client := &http.Client{}
					//					client.Do(req)

				}
			}
		}
	}
}

func main() {

	syslog, err := syslog.New(syslog.LOG_INFO, "windmaker-alarmsensors")
	if err != nil {
		panic(err)
	}

	syslog.Info("Reading service config.")
	serviceConfig, errConfig := config.ReadConfig()

	if errConfig != nil {
		panic(errConfig)
	}

	httpClient := http.Client{
		Timeout: time.Second * 5, // Maximum of 5 secs
	}

	syslog.Info("Establishing connection with alarmManager.")

	watcher := apiwatcher.APIWatcher{Host: serviceConfig.AlarmManager.Host, Port: serviceConfig.AlarmManager.Port}
	alarmManagerRequester := apiwatcher.Requester{Client: httpClient}

	_, apiInfoErr := watcher.ShowInfo(alarmManagerRequester)
	if apiInfoErr != nil {
		errorString := fmt.Sprintf("%v", apiInfoErr.Error())
		syslog.Err(errorString)
		panic(apiInfoErr)
	}

	mqttMessages := make(chan [2]string)
	syslog.Info("Establishing connection with mqtt server.")
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", serviceConfig.Mqtt.Host, serviceConfig.Mqtt.Port))
	opts.SetClientID("windmaker_alarmsensors")
	opts.SetUsername(serviceConfig.Mqtt.User)
	opts.SetPassword(serviceConfig.Mqtt.Password)

	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		mqttMessages <- [2]string{msg.Topic(), string(msg.Payload())}
	})

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		errorString := fmt.Sprintf("%v", token.Error())
		syslog.Err(errorString)
		panic(token.Error())
	}
	sub(client, serviceConfig.Mqtt.WildcardTopic, syslog)

	syslog.Info("Connection established.")

	for {
		incoming := <-mqttMessages
		go handleMessage(serviceConfig, syslog, watcher, alarmManagerRequester, incoming[0], incoming[1])
	}

}
