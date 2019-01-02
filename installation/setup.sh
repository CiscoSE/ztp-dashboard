#!/bin/bash

#install golang
sudo apt-update

wget https://dl.google.com/go/go1.11.4.linux-amd64.tar.gz
sudo tar -xvf go1.11.4.linux-amd64.tar.gz
sudo mv go /usr/local

#setup go files
mkdir $HOME/asic_q2
cd $HOME/asic_q2
mkdir src 
mkdir bin
mkdir pkg
go get github.com/CiscoSE/ztp-dashboard
go install github.com/CiscoSE/ztp-dashboard

#install isc-dhcp-server
sudo apt update
sudo apt install isc-dhcp-server

#start service
sudo systemctl start isc-dhcp-server.service
sudo systemctl start isc-dhcp-server.service

#install tftp server
sudo apt update
sudo apt install xinetd tftpd tftp

#create tftpboot directory
sudo mkdir /tftpboot
sudo chmod -R 777 /tftpboot
sudo chown -R nobody /tftpboot

#start tftp
sudo systemctl start xinetd

#run ztp-dashboard as a daemon
cd $HOME/asic_q2/bin
sudo daemonize -p /var/run/ztp-dashboard.pid -l /var/lock/subsys/ztp-dashboard -u nobody $HOME/asic_q2/bin/ztp-dashboard

