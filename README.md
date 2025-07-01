# MeshSpy

MeshSpy reads packets from a Meshtastic device over a serial port and publishes
nodes information to an MQTT broker. The project includes a Go binary and a
Dockerfile to build a minimal container image.

## Requirements

- Go 1.20+ for building locally
- `meshtastic-go` binary available in `/usr/local/bin/meshtastic-go` or built
  using the provided Dockerfile
- A `.env.runtime` file with runtime settings

## Building

Create a `.env.build` from the provided example:

```bash
cp .env.build.example .env.build
# edit .env.build with your Docker credentials
```

Then build the image:

```bash
./build.sh
```

## Running

Configure the runtime environment in `.env.runtime`. The most important
variable is `SERIAL_PORT` which should point to your Meshtastic serial device.
Then run the container exposing the serial device and MQTT details:

```bash
docker run --device=/dev/ttyACM0 \
  --env-file .env.runtime nicbad/meshspy
```

During start-up the service prints information from `meshtastic-go` and begins
streaming data from the serial port to the configured MQTT topic.
