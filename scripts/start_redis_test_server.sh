#!/bin/bash -
#===============================================================================
#
#          FILE: start_redis_test_server.sh
#
#         USAGE: ./start_redis_test_server.sh
#
#   DESCRIPTION: Script made in order to manager redis docker image used
#                during integration tests.
#
#  REQUIREMENTS: User must have sudo privileges
#        AUTHOR: Álvaro Castellano Vela (alvaro.castellano.vela@gmail.com),
#       CREATED: 06/06/2022 07:59
#===============================================================================

set -o nounset                              # Treat unset variables as an error

# Remove existing images

docker stop $(docker ps -a --filter name=redis_alarmsensors_test_server -q) 2> /dev/null > /dev/null
docker rm $(docker ps -a --filter name=redis_alarmsensors_test_server -q) 2> /dev/null > /dev/null

# Create docker image

docker create --name redis_alarmsensors_test_server -p 6379:6379 registry.windmaker.net:5005/a-castellano/limani/base_redis_server 2> /dev/null > /dev/null

docker start redis_alarmsensors_test_server > /dev/null
