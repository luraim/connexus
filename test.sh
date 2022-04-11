#!/bin/bash
export BIN=dist/connexusserver
./build.sh &&\
$BIN testwiki test1 46723
