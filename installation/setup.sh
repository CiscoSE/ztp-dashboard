#!/bin/bash

# Create the following .env file with the correct variables for your setup
#--------------------------------------------------------------------------
# Go related variables, shouldn't need to be changed
#export GOPATH=/opt/ztp-go
#export GOBIN=$GOPATH/bin
#export GOROOT=/usr/local/go
#export PATH=$PATH:$GOPATH/bin

# DHCP v4 information
#export DHCP_NAMESERVERS=
#export DHCP_SUBNET=
#export DHCP_SUBNET_NETMASK=
#export DHCP_CONFIG_PATH=/etc/dhcp/dhcpd.conf
#export DHCP_SERVICE_RESTART_CMD="systemctl restart isc-dhcp-server"

# DHCP v6 information
#export DHCP6_NAMESERVERS=
#export DHCP6_SUBNET=
#export DHCP6_SUBNET_NETMASK=
#export DHCP6_CONFIG_PATH=/etc/dhcp/dhcpd6.conf
#export DHCP6_SERVICE_RESTART_CMD="systemctl restart isc-dhcp-server6"

# Mongo URI to be used by the tool
#export DB_URI=
# Port to be listening for incomming web requests
#export APP_WEB_PORT=8080
# Token to be used when sending notifications
#export WEBEX_BOT_TOKEN=
# Enable for extra log information
#export DEBUG=on
#--------------------------------------------------------------------------


#install golang
sudo apt update
wget https://dl.google.com/go/go1.11.4.linux-amd64.tar.gz
sudo tar -xvf go1.11.4.linux-amd64.tar.gz
sudo mv go /usr/local
echo 'export GOROOT=/usr/local/go' | sudo tee -a /etc/profile
echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee -a /etc/profile
source /etc/profile

#setup go files
mkdir /opt/ztp-go
cp env /opt/ztp-go/.env
cd /opt/ztp-go
sudo chmod 755 .env
. .env
mkdir src
mkdir bin
mkdir pkg
go get github.com/CiscoSE/ztp-dashboard
go install github.com/CiscoSE/ztp-dashboard

#install isc-dhcp-server
sudo apt install -y isc-dhcp-server

#start service
sudo systemctl start isc-dhcp-server.service
sudo systemctl start isc-dhcp-server6.service

#install tftp server

sudo apt install -y xinetd tftpd tftp

#create tftpboot directory
sudo mkdir /tftpboot
sudo chmod -R 777 /tftpboot
sudo chown -R nobody /tftpboot

tftp_file="/etc/xinetd.d/tftp"

echo "service tftp" >> $tftp_file
echo "{" >> $tftp_file
echo "protocol        = udp" >> $tftp_file
echo "port            = 69" >> $tftp_file
echo "socket_type     = dgram" >> $tftp_file
echo "wait            = yes" >> $tftp_file
echo "user            = nobody" >> $tftp_file
echo "server          = /usr/sbin/in.tftpd" >> $tftp_file
echo "server_args     = /tftpboot" >> $tftp_file
echo "disable         = no" >> $tftp_file
echo "}" >> $tftp_file

#start tftp
sudo systemctl start xinetd

#run ztp-dashboard as a daemon
#create a system file for ztp-dashboard
system_file="/etc/systemd/system/ztp-dashboard.service"

echo "[Unit]" >> $system_file
echo "Description=ZTP-dashboard service" >> $system_file
echo "[Service]" >> $system_file
echo "ExecStart=/opt/ztp-go/bin/start-ztp.sh" >> $system_file
echo "[Install]" >> $system_file
echo "WantedBy=multi-user.target" >> $system_file


#create a startup file for ztp-dashboard
cd /opt/ztp-go/bin
startup_file="start-ztp.sh"

echo "#!/bin/bash" >> $startup_file
echo "cd /opt/ztp-go" >> $startup_file
echo ". .env" >> $startup_file
echo "cd /opt/ztp-go/bin/" >> $startup_file
echo "./ztp-dashboard" >> $startup_file

sudo chmod 755 $startup_file
sudo systemctl start ztp-dashboard

#link created so config and images can be accessed via tftp
ln -sf $GOPATH/src/github.com/CiscoSE/ztp-dashboard/public/ /tftpboot/
