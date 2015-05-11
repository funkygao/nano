#!/bin/sh 

rm -f pipeline
go build

./pipeline pull tcp://127.0.0.1:1235 & pull=$!
./pipeline push tcp://127.0.0.1:1234 & push=$!
./pipeline forward tcp://127.0.0.1:1234 tcp://127.0.0.1:1235 & forward=$!
sleep 5
kill $push $pull $forward
