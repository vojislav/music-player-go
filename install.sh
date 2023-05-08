#!/bin/bash

mkdir -p $HOME/.config/music-player-go
cp init.sql $HOME/.config/music-player-go
go build .