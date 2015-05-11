#!/bin/sh 

rm -f reqrep
go build

./reqrep req tcp://127.0.0.1:1234 & client=$!
./reqrep rep tcp://127.0.0.1:1234 & server=$!
sleep 5
kill $server $client
