package main

import (
	"bytes"
	"fmt"
	"log/syslog"
	"net/http"
	"time"

	alarmsensors "github.com/a-castellano/AlarmSensors/alarmsensors"
	config "github.com/a-castellano/AlarmSensors/config_reader"
	storage "github.com/a-castellano/AlarmSensors/storage"
	apiwatcher "github.com/a-castellano/AlarmStatusWatcher/apiwatcher"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	goredis "github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
	"golang.org/x/net/context"
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

func sendMessageByQueue(rabbitmqConfig config.Rabbitmq, messageToSend string) error {

	dialString := fmt.Sprintf("amqp://%s:%s@%s:%d/", rabbitmqConfig.User, rabbitmqConfig.Password, rabbitmqConfig.Host, rabbitmqConfig.Port)
	conn, errDial := amqp.Dial(dialString)
	defer conn.Close()

	if errDial != nil {
		fmt.Println(errDial)
		return errDial
	}

	channel, errChannel := conn.Channel()
	defer channel.Close()
	if errChannel != nil {
		return errChannel
	}

	queue, errQueue := channel.QueueDeclare(
		rabbitmqConfig.Queue, // name
		true,                 // durable
		false,                // delete when unused
		false,                // exclusive
		false,                // no-wait
		nil,                  // arguments
	)
	if errQueue != nil {
		return errQueue
	}

	// send Job

	err := channel.Publish(
		"",         // exchange
		queue.Name, // routing key
		false,      // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(messageToSend),
		})

	if err != nil {
		return err
	}
	return nil

}

func handleMessage(ctx context.Context, serviceConfig config.Config, syslog *syslog.Writer, watcher apiwatcher.APIWatcher, alarmManagerRequester apiwatcher.Requester, topic string, message string, storageInstance storage.Storage) {

	candidateSensor := alarmsensors.RetriveChildTopic(topic, serviceConfig.Mqtt.WildcardTopic)

	if _, sensorIsManaged := serviceConfig.Sensors[candidateSensor]; sensorIsManaged {
		statusMessage, sensorActivated, checkSensorErr := alarmsensors.CheckSensorTriggered(ctx, candidateSensor, message, storageInstance)
		if checkSensorErr != nil {
			errorString := fmt.Sprintf("%v", checkSensorErr.Error())
			syslog.Err(errorString)
		} else {
			syslog.Info(statusMessage)
			// Check alarm status
			if sensorActivated == true {
				apiInfo, apiInfoErr := watcher.ShowInfo(alarmManagerRequester)
				if apiInfoErr != nil {
					apiErrorString := fmt.Sprintf("%v", apiInfoErr.Error())
					syslog.Err(apiErrorString)
				} else {
					currentAlarmMode := apiInfo.DevicesInfo[serviceConfig.AlarmManager.DeviceId].Mode
					// Check if sensor triggers alarm
					if _, triggerAlarm := serviceConfig.Sensors[candidateSensor].SensorTriggers[currentAlarmMode]; triggerAlarm {
						logMessage := fmt.Sprintf("%s sensor has been triggered and alarm status is %s, triggering alarm.", candidateSensor, currentAlarmMode)
						syslog.Info(logMessage)
						sendMessageByQueue(serviceConfig.Rabbitmq, logMessage)
						jsonString := fmt.Sprintf("{\"mode\":\"SOS\"}")
						var jsonStr = []byte(jsonString)
						apiURL := fmt.Sprintf("http://%s:%d/devices/status/%s", serviceConfig.AlarmManager.Host, serviceConfig.AlarmManager.Port, serviceConfig.AlarmManager.DeviceId)
						req, _ := http.NewRequest("PUT", apiURL, bytes.NewBuffer(jsonStr))
						req.Header.Set("Content-Type", "application/json")
						client := &http.Client{}
						client.Do(req)
					} else {
						logMessage := fmt.Sprintf("DEBUG - %s sensor has been triggered but alarm status is %s, NOT triggering alarm.", candidateSensor, currentAlarmMode)
						syslog.Info(logMessage)
						sendMessageByQueue(serviceConfig.Rabbitmq, logMessage)
					}
				}
			} else {
				debugMessage := fmt.Sprintf("DEBUG - %s", statusMessage)
				syslog.Info(debugMessage)
				sendMessageByQueue(serviceConfig.Rabbitmq, debugMessage)
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

	redisAddress := fmt.Sprintf("%s:%d", serviceConfig.RedisServer.IP, serviceConfig.RedisServer.Port)
	redisClient := goredis.NewClient(&goredis.Options{
		Addr:     redisAddress,
		Password: serviceConfig.RedisServer.Password,
		DB:       serviceConfig.RedisServer.Database,
	})

	ctx := context.Background()

	redisErr := redisClient.Set(ctx, "checkKey", "key", 1000000).Err()
	if redisErr != nil {
		panic(redisErr)
	}
	storageInstance := storage.Storage{RedisClient: redisClient}

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
		go handleMessage(ctx, serviceConfig, syslog, watcher, alarmManagerRequester, incoming[0], incoming[1], storageInstance)
	}

}
