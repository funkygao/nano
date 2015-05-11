#!/bin/sh 

rm -f pubsub
go build

./pubsub sub & sub=$!
./pubsub pub & pub=$!
sleep 5
kill $pub $sub
