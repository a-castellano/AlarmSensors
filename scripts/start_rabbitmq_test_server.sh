#!/bin/bash -
#===============================================================================
#
#          FILE: start_rabbitmq_test_server.sh
#
#         USAGE: ./start_rabbitmq_test_server.sh
#
#   DESCRIPTION: Script made in order to manager rabbitmq docker image used
#                during integration tests.
#
#  REQUIREMENTS: User must have sudo privileges
#        AUTHOR: Álvaro Castellano Vela (alvaro.castellano.vela@gmail.com),
#       CREATED: 09/12/2020 17:29
#===============================================================================

set -o nounset                              # Treat unset variables as an error

# Remove existing images

docker stop $(docker ps -a --filter name=rabbitmq_metal_archives_wrapper_test_server -q) 2> /dev/null > /dev/null
docker rm $(docker ps -a --filter name=rabbitmq_metal_archives_wrapper_test_server -q) 2> /dev/null > /dev/null

# Create docker image

docker create --name rabbitmq_alarm_sensors_test_server -p 5672:5672 -p 15672:15672 registry.windmaker.net:5005/a-castellano/limani/base_rabbitmq_server 2> /dev/null > /dev/null

docker start rabbitmq_alarm_sensors_test_server > /dev/null
