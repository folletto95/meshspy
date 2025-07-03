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

### Building binaries

To produce standalone binaries for Linux, Windows and macOS, run:

```bash
./build-binaries.sh
```

The compiled executables will be placed in the `dist/` directory.


## Running

Configure the runtime environment in `.env.runtime`. The most important
variable is `SERIAL_PORT` which should point to your Meshtastic serial device.
Then run the container exposing the serial device and MQTT details:

```bash
docker run --device=/dev/ttyACM0 \
  --env-file .env.runtime nicbad/meshspy
```

During start-up the service prints information from `meshtastic-go` and begins

streaming data from the serial port to the configured MQTT topic. When the
`SEND_ALIVE_ON_START` environment variable is set to `true`, the service also
sends a `MeshSpy Alive` message on the configured MQTT topic and to the node
itself using `meshtastic-go --sendtext`, so other components can detect that
the service is running and nodes are reached.


### `start_meshspy.sh` helper

For a quick start, run the `start_meshspy.sh` script which launches the
container with some default environment variables:

- `SERIAL_PORT=/dev/ttyACM0`
- `BAUD_RATE=115200`
- `MQTT_BROKER=tcp://smpisa.ddns.net:1883`
- `MQTT_TOPIC=meshspy`
- `MQTT_CLIENT_ID=meshspy-kali`
- `MQTT_USER=testmeshspy`
- `MQTT_PASS=test1`
- `SEND_ALIVE_ON_START=false`
  (set to `true` if you want the service to send and log a `MeshSpy Alive`
  message on start-up)

Start the container using the defaults:

```bash
./start_meshspy.sh
```

Use `--clean` to remove and re-pull the image before starting and `--log` to
periodically show container logs:

```bash
./start_meshspy.sh --clean
./start_meshspy.sh --log
```

Both options can be combined if required.

## Web Application

A simple web interface lives in `cmd/webapp`. It serves an HTML page and
forwards MQTT messages over WebSockets. Run it with Go:

```bash
go run ./cmd/webapp
```

The application reads the same `.env.runtime` file used by `meshspy`. Set
`WEB_PORT` to change the listening port (default `8080`) and open your browser
to `http://localhost:8080`.

## Simple Message Board

For a minimal example that does not rely on MQTT, a tiny in-memory
message board is available in `cmd/messagesapp`:

```bash
go run ./cmd/messagesapp
```

Visit `http://localhost:8080` and post messages through the form to see
them listed on the page.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.