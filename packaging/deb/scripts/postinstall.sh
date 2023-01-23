#!/bin/sh

mkdir -p /etc/windmaker-alarmsensors

echo "### NOT starting on installation, please execute the following statements to configure windmaker-alarmsensors to start automatically using systemd"
echo " sudo /bin/systemctl daemon-reload"
echo " sudo /bin/systemctl enable windmaker-alarmsensors"
echo "### You can start grafana-server by executing"
echo " sudo /bin/systemctl start windmaker-alarmsensors"
