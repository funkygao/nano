#!/bin/sh 

rm -f pair
go build

./pair server & server=$!
./pair client & client=$!
sleep 5
kill $server $client
