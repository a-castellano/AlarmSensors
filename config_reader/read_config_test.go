package config

import (
	"os"
	"testing"
)

func TestProcessConfigNoMqtt(t *testing.T) {
	os.Setenv("ALARM_SENSORS_CONFIG_FILE_LOCATION", "./config_files_test/config_no_mqtt/")
	_, err := ReadConfig()
	if err == nil {
		t.Errorf("ReadConfig method without mqtt should fail.")
	} else {
		if err.Error() != "Fatal error config: no mqtt field was found." {
			t.Errorf("Error should be \"Fatal error config: no mqtt field was found.\" but error was '%s'.", err.Error())
		}
	}
}

func TestProcessConfigNoRabbitmqHost(t *testing.T) {
	os.Setenv("ALARM_SENSORS_CONFIG_FILE_LOCATION", "./config_files_test/config_no_rabbitmq_host/")
	_, err := ReadConfig()
	if err == nil {
		t.Errorf("ReadConfig method without rabbitmq host should fail.")
	} else {
		if err.Error() != "Fatal error config: no rabbitmq host was found." {
			t.Errorf("Error should be \"Fatal error config: no rabbitmq host was found.\" but error was '%s'.", err.Error())
		}
	}
}

func TestProcessConfigNoRedisConfig(t *testing.T) {
	os.Setenv("ALARM_SENSORS_CONFIG_FILE_LOCATION", "./config_files_test/config_no_redis/")
	_, err := ReadConfig()
	if err == nil {
		t.Errorf("ReadConfig method without redis config should fail.")
	} else {
		if err.Error() != "Fatal error config: no redis ip was defined." {
			t.Errorf("Error should be \"Fatal error config: no redis ip was defined.\" but error was '%s'.", err.Error())
		}
	}
}

func TestProcessConfigDuplicatedTriggers(t *testing.T) {
	os.Setenv("ALARM_SENSORS_CONFIG_FILE_LOCATION", "./config_files_test/config_duplicated_trigger/")
	_, err := ReadConfig()
	if err == nil {
		t.Errorf("ReadConfig method with duplicated triggers should fail.")
	} else {
		if err.Error() != "Fatal error reading config file: While parsing config: toml: table home_armed already exists" {
			t.Errorf("Error should be \"Fatal error reading config file: While parsing config: toml: table home_armed already exists\" but error was '%s'.", err.Error())
		}
	}
}

func TestOKConfig(t *testing.T) {
	os.Setenv("ALARM_SENSORS_CONFIG_FILE_LOCATION", "./config_files_test/config_ok/")
	config, err := ReadConfig()
	if err != nil {
		t.Errorf("ReadConfig with ok config shouln't return errors. Returned: %s.", err.Error())
	}
	if config.Mqtt.Host != "localhost" {
		t.Errorf("Mqtt Mqtt should be localhost. Returned: %s.", config.Mqtt.Host)
	}
	if len(config.SensorTriggers) != 2 {
		t.Errorf("SensorTriggers length should be 2. Returned: %d.", len(config.SensorTriggers))
	}
	if config.SensorTriggers["home_armed"].Name != "home_armed" {
		t.Errorf("SensorTrigger home_armed name should be 'home_armed' Returned: %s.", config.SensorTriggers["home_armed"].Name)
	}
	doorSensor := config.SensorTriggers["home_armed"].Sensors["door1"]
	if doorSensor.Name != "door1" {
		t.Errorf("doorSensor Name should be door1. Returned: %s.", doorSensor.Name)
	}
}
