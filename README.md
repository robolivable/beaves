## Beaves

Beacon Jeeves (Beaves) is a BLE-based proximity sentry that can manage switches over Raspberry Pi's GPIO. It uses `GPIO17` and `GPIO27` as a backup.

#### iPhone users

Beaves doesn't work with iPhones with enabled MAC address randomization. If you want to use an iPhone with Beaves you must disable this feature.

### Dependencies

On Raspbian, you mainly need `bluez`, its debugging tools (used to facilitate pairing), and development files:

```sh
apt install bluez bluez-tools libbluetooth-dev
```

### BLE agent service

This app depends on bt-agent utils to facilitate pairing. `start-agent.sh` manages the agent. Configure and install the `.service` files (using `systemctl`) in this project to run both the agent and Beaves on a Raspberry Pi.

When fully installed, Beaves runs autonomously on boot. It's designed to run forever and forget all paired devices on reboot. If your device stops pairing, it's likely the Pi restarted. In this case, simply "forget" the sentry on your device and re-pair it.

##### Using a BLE dongle with Raspi

Make sure you disable onboard bluetooth and the UART service that manages it:

```sh
echo "dtoverlay=disable-bt" >> /boot/firmware/config.txt
systemctl disable hciuart
```

### Build

It's easier to cross-compile for Raspbian instead of building on the Rasp Pi itself since it's way faster, and you get access to the most recent version of Golang and its modules:

```sh
git clone https://github.com/robolivable/beaves.git && cd beaves && go mod tidy
GOOS=linux GOARCH=arm64 go build -o beaves
```

I use `scp` to copy the executable afterward:

```sh
scp beaves rob@192.168.1.42:/home/raspi
```

### Config

A `config.json` file is required at runtime in the working directory. E.g.:

```json
{
  "bluetooth": {
    "advertisementName": "Beaves Sentry",
    "advertisementDelayMs": 30000,
    "connectionPoolSize": 10,
    "connectionsLimit": 1,
    "connectionLimitDelayMs": 20000,
    "disconnectionDelayMs": 3000
  },
  "actors": {
    "known": [
      "11:22:33:AA:BB:CC"
    ]
  },
  "log": {
    "enabled": true,
    "debug": false
  },
  "eventLoopDelayMs": 3000,
  "operationDelayMs": 30000
}
```

I suggest using the above delay values and setting your BT device's MAC address (i.e., your phone).

### License

MIT
