#!/bin/bash
export BIN=connexusserver
rm -rf dist
mkdir dist
packr2 && \
go build -o dist/$BIN main.go && \
cp dist/$BIN $HOME/local/bin/ &&\
packr2 clean
