#!/bin/bash

tc qdisc add dev lo:1 root netem delay 100ms 
tc qdisc change dev lo:1 root netem delay 100ms 20ms distribution normal
tc qdisc show dev lo:1
