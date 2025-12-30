#!/bin/bash
bluetoothctl devices | awk '{print $2}' | xargs -I{} bluetoothctl remove {} && bt-agent -c NoInputNoOutput