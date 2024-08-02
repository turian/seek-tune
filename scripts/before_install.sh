#!/usr/bin/env bash

sudo apt-get -y update

# install golang
sudo apt-get -y install golang-go

# install nodeJS and npm
sudo apt -y install nodejs
sudo apt -y install npm

# install ffmpeg
sudo apt-get -y install ffmpeg

# install Certbot
DOMAIN="localport.online"
EMAIL="cgzirim@gmail.com"
CERT_DIR="/etc/letsencrypt/live/$DOMAIN"

if [ ! -f "$CERT_DIR" ]; then
    sudo apt install -y certbot
    sudo certbot certonly --standalone -d $DOMAIN --email $EMAIL --agree-tos --non-interactive
    if [ $? -eq 0 ]; then
        sudo apt-get -y install acl
        sudo setfacl -m u:ubuntu:--x /etc/letsencrypt/archive
  fi
fi

sudo rm -rf /home/ubuntu/song-recognition
