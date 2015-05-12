#!/bin/sh 

rm -f reqrep
go build

./reqrep req1 tcp://127.0.0.1:1234 & client1=$!
./reqrep req2 tcp://127.0.0.1:1234 & client2=$!
./reqrep rep tcp://127.0.0.1:1234 & server=$!
sleep 1
kill $server $client1 $client2
