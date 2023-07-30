package config

import (
	"errors"

	viperLib "github.com/spf13/viper"
)

type Mqtt struct {
	Host          string
	Port          int
	User          string
	Password      string
	WildcardTopic string
}

type Rabbitmq struct {
	Host     string
	Port     int
	User     string
	Password string
	Queue    string
}

type AlarmManager struct {
	Host     string
	Port     int
	DeviceId string
}

type Sensor struct {
	Name           string
	SensorTriggers map[string]bool
}

type SensorTrigger struct {
	Name    string
	Sensors map[string]*Sensor
}

type RedisServer struct {
	IP       string
	Port     int
	Password string
	Database int
}

type Config struct {
	Mqtt           Mqtt
	Rabbitmq       Rabbitmq
	AlarmManager   AlarmManager
	Sensors        map[string]*Sensor
	SensorTriggers map[string]SensorTrigger
	RedisServer    RedisServer
}

func ReadConfig() (Config, error) {
	var configFileLocation string
	var config Config

	var envVariable string = "ALARM_SENSORS_CONFIG_FILE_LOCATION"

	requiredVariables := []string{"mqtt", "sensor_triggers", "rabbitmq", "alarmmanager"}
	mqttRequiredVariables := []string{"host", "port", "user", "password", "wildcard_topic"}
	rabbitmqRequiredVariables := []string{"host", "port", "user", "password", "queue"}
	alarmManagerRequiredVariables := []string{"host", "port", "deviceid"}
	redisRequiredVariables := []string{"ip", "port", "password", "database"}

	viper := viperLib.New()

	//Look for config file location defined as env var
	viper.BindEnv(envVariable)
	configFileLocation = viper.GetString(envVariable)
	if configFileLocation == "" {
		// Get config file from default location
		return config, errors.New(errors.New("Environment variable SECURITY_CAM_BOT_CONFIG_FILE_LOCATION is not defined.").Error())
	}

	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(configFileLocation)

	if err := viper.ReadInConfig(); err != nil {
		return config, errors.New(errors.New("Fatal error reading config file: ").Error() + err.Error())
	}

	for _, requiredVariable := range requiredVariables {
		if !viper.IsSet(requiredVariable) {
			return config, errors.New("Fatal error config: no " + requiredVariable + " field was found.")
		}
	}

	for _, mqttVariable := range mqttRequiredVariables {
		if !viper.IsSet("mqtt." + mqttVariable) {
			return config, errors.New("Fatal error config: no mqtt " + mqttVariable + " was found.")
		}
	}

	for _, rabbitmqVariable := range rabbitmqRequiredVariables {
		if !viper.IsSet("rabbitmq." + rabbitmqVariable) {
			return config, errors.New("Fatal error config: no rabbitmq " + rabbitmqVariable + " was found.")
		}
	}

	for _, alarmManagerVariable := range alarmManagerRequiredVariables {
		if !viper.IsSet("alarmmanager." + alarmManagerVariable) {
			return config, errors.New("Fatal error config: no alarmManager " + alarmManagerVariable + " was found.")
		}
	}

	readedSenorTriggerNames := make(map[string]bool)
	readedSensorNames := make(map[string]bool)

	sensors := make(map[string]*Sensor)
	sensorTriggers := make(map[string]SensorTrigger)
	//

	readedSenorTriggers := viper.GetStringMap("sensor_triggers")

	for readedSenorTriggerName := range readedSenorTriggers {
		if _, ok := readedSenorTriggerNames[readedSenorTriggerName]; ok {
			return config, errors.New("Fatal error config: sensor trigger called " + readedSenorTriggerName + " was already declared.")
		} else {
			readedSenorTriggerNames[readedSenorTriggerName] = true
		}

		newSensorTrigger := SensorTrigger{Name: readedSenorTriggerName}
		newSensorTrigger.Sensors = make(map[string]*Sensor)
		sensorTriggers[readedSenorTriggerName] = newSensorTrigger

		sensorList := viper.GetStringSlice("sensor_triggers." + readedSenorTriggerName + ".sensors")
		for _, sensorName := range sensorList {
			if _, ok := readedSensorNames[sensorName]; !ok {
				readedSensorNames[sensorName] = true
				newSensor := Sensor{Name: sensorName}
				newSensor.SensorTriggers = make(map[string]bool)
				sensors[sensorName] = &newSensor
			}
			sensors[sensorName].SensorTriggers[readedSenorTriggerName] = true
			sensorTriggers[readedSenorTriggerName].Sensors[sensorName] = sensors[sensorName]

		}
		//Check if sensor already exists
	}

	rabbitmqConfig := Rabbitmq{Host: viper.GetString("rabbitmq.host"), Port: viper.GetInt("rabbitmq.port"), User: viper.GetString("rabbitmq.user"), Password: viper.GetString("rabbitmq.password"), Queue: viper.GetString("rabbitmq.queue")}

	mqttConfig := Mqtt{Host: viper.GetString("mqtt.host"), Port: viper.GetInt("mqtt.port"), User: viper.GetString("mqtt.user"), Password: viper.GetString("mqtt.password"), WildcardTopic: viper.GetString("mqtt.wildcard_topic")}

	alarmManagerConfig := AlarmManager{Host: viper.GetString("alarmmanager.host"), Port: viper.GetInt("alarmmanager.port"), DeviceId: viper.GetString("alarmmanager.deviceid")}

	config.Rabbitmq = rabbitmqConfig
	config.Mqtt = mqttConfig
	config.AlarmManager = alarmManagerConfig

	config.Sensors = sensors
	config.SensorTriggers = sensorTriggers

	// Redis
	for _, requiredRedisVariable := range redisRequiredVariables {
		if !viper.IsSet("redis." + requiredRedisVariable) {
			return config, errors.New("Fatal error config: no redis " + requiredRedisVariable + " was defined.")
		}

	}
	config.RedisServer.IP = viper.GetString("redis.ip")
	config.RedisServer.Port = viper.GetInt("redis.port")
	config.RedisServer.Password = viper.GetString("redis.password")
	config.RedisServer.Database = viper.GetInt("redis.database")

	return config, nil
}
