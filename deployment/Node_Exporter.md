# Monitoring server itself

## Download node_exporter

## Run as service

```
$ sudo vim /etc/systemd/system/node_exporter.service
[Unit]
Description=Node Exporter

[Service]
User=ubuntu
ExecStart=/home/ubuntu/Prometheus/node_exporter/node_exporter

[Install]
WantedBy=default.target
```

```
 sudo systemctl daemon-reload
 sudo systemctl enable node_exporter.service
 sudo systemctl start node_exporter.service
 sudo systemctl status node_exporter.service
```

https://vexxhost.com/resources/tutorials/how-to-use-prometheus-to-monitor-your-centos-7-server/