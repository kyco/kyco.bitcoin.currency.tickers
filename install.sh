#!/bin/bash

# Clear the current cli screen
clear

# # Check root access to copy over the systemd files
# # as well as place the binary in /usr/bin
# if [[ $EUID -ne 0 ]]; then
#    echo "This script must be run as root" 
#    exit 1
# fi

# Explain to the user that we are about to install
# some files etc
echo "Welcome to the bitcoin currency ticker golang service."
sleep 2
echo "This install script will compile the binary using golang."
sleep 2
echo "Checking if golang is installed..."
sleep 2

# Check if golang is currently installed
command -v go >/dev/null 2>&1 || { echo >&2 "I require go but it's not installed. Aborting."; exit 1;}
# If we didn't exit, we found go
echo "Go is installed"
sleep 2

# Check if glide is installed, if not exit
command -v glide >/dev/null 2>&1 || { echo >&2 "I require glide but it's not installed. Aborting."; exit 1;}
# If we didn't exit, we found glide
echo "Glide is installed"
sleep 2

echo "Installing dependencies..."
sleep 2

# Install dependencies
glide install

# Inform user of compile
echo "Compiling binary..."
sleep 2

# Actually compile it
go build

echo "Binary compiled!"
sleep 2

# Copy the binary to /usr/bin
echo "Copying binary to /usr/bin/"
sleep 2
echo "Please enter root password to copy binary to /usr/bin/"
sudo mv kyco.bitcoin.currency.tickers /usr/bin/

# Figure out user and group
iam=$(whoami)
group=$(id -g -n $iam)
cp init/kbct.service /tmp
sed -i -e 's/User=user/User='$iam'/g' /tmp/kbct.service
sed -i -e 's/Group=user/Group='$group'/g' /tmp/kbct.service

# Copy the service file to /usr/lib/systemd/user/
echo "Copying service file to /usr/lib/systemd/user/"
sleep 2
sudo mv /tmp/kbct.service /usr/lib/systemd/user/

# Configure systemd with new service file
echo "Configure systemd"
sleep 2
sudo systemctl enable /usr/lib/systemd/user/kbct.service; sudo systemctl daemon-reload
sudo service kbct start

echo "Done! Run the service using 'service kbct start | restart | status | stop'"