#!/bin/bash

export PATH=$PATH:/usr/local/go/bin

go version

tmux kill-server

rm -rf icapeg

tmux new-session -s icapeg -d

git clone "https://github.com/egirna/icapeg"

cd /home/linux/icapeg

rm config.toml
cp /home/linux/config.toml .

git checkout develop

git branch

go run main.go
