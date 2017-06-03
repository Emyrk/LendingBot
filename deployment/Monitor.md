# Monitoring server

 - Prometheus
 - Grafana
 - Node Exporter

## Download Prometheus

```
mkdir -p ~/Prometheus
wget https://github.com/prometheus/prometheus/releases/download/v2.0.0-alpha.2/prometheus-2.0.0-alpha.2.linux-amd64.tar.gz
tar -xvzf prometheus-2.0.0-alpha.2.linux-amd64.tar.gz
mv prometheus-2.0.0-alpha.2.linux-amd64.tar.gz ~/Prometheus/server
```

## Install Node Exporter

```
mkdir -p ~/Prometheus/node_exporter
cd ~/Prometheus/node_exporter
wget https://github.com/prometheus/node_exporter/releases/download/v0.14.0/node_exporter-0.14.0.linux-amd64.tar.gz
tar -xvzf node_exporter-0.14.0.linux-amd64.tar.gz 
mv node_exporter-0.14.0.linux-amd64/* .
sudo ln -s ~/Prometheus/node_exporter/node_exporter /usr/bin
```

See Node_Exporter.md for running instructions

## Run Prometheus

```
cd /home/ubuntu/Prometheus/server
nohup ./prometheus > prometheus.log 2>&1 &
```


## Download Grafana 

http://docs.grafana.org/installation/debian/

## Run Grafana

```
sudo service grafana-server start
```