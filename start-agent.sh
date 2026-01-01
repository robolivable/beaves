#!/bin/bash
rfkill unblock bluetooth
sleep 2
hciconfig hci0 up
bluetoothctl power on
pkill bt-agent
bluetoothctl devices | awk '{print $2}' | xargs -I{} bluetoothctl remove {}
bt-agent -c NoInputNoOutput