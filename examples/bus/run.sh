#!/bin/sh 

url0=tcp://127.0.0.1:40890
url1=tcp://127.0.0.1:40891
url2=tcp://127.0.0.1:40892
url3=tcp://127.0.0.1:40893
./bus node0 $url0 $url1 $url2 & node0=$!
./bus node1 $url1 $url2 $url3 & node1=$!
./bus node2 $url2 $url3 & node2=$!
./bus node3 $url3 $url0 & node3=$!
sleep 5
kill $node0 $node1 $node2 $node3

