#!/bin/sh 

go build
./benchmark pull & pull=$!
./benchmark push & push=$1
sleep 21600 # 6h
kill $push $pull
