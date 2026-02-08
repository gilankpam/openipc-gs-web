# OpenIPC EZConfig API

This is a lightweight Go-based API service designed to run on the OpenIPC Air Unit. It provides a RESTful interface to configure key services without manual file editing.

## Features

- **Radio Configuration**: Update WFB-ng settings (Frequency, Power, Bandwidth).
- **Video Configuration**: Update Majestic encoding settings (Resolution, FPS, Bitrate).
- **Camera Configuration**: Update sensor settings (Exposure, Flip, Mirror).
- **Adaptive Link**: Enable/Disable adaptive link logic and configure parameters.
- **Automatic Service Management**: Automatically restarts relevant services (`wifibroadcast`, `majestic`, `alink`) when configurations are changed.
- **Persistence**: Updates are saved to the standard OpenIPC configuration files (`/etc/wfb.yaml`, `/etc/majestic.yaml`, `/etc/alink.conf`, `/etc/rc.local`).

## Building for Air Unit

The OpenIPC Air Unit typically runs on ARM architecture. Cross-compile the binary:

```bash
make build-air-unit
scp -O ezconfig root@X.X.X.X:/usr/bin/ezconfig
scp -O conf/S99ezconfig root@X.X.X.X:/etc/init.d/
```

Reboot the Air Unit to apply the changes:

```bash
reboot
```

## Ground Station (GS) WebUI

The EZConfig system includes a Ground Station WebUI that acts as a frontend for the Air Unit API. It runs on the Ground Station (e.g., Raspberry Pi, Laptop) and proxies requests to the Air Unit.

### 1. Build the Frontend

Requires Node.js and npm.

```bash
cd web
npm install
npm run build
```

This will generate static files in `web/dist`.

### 2. Build the GS Server

```bash
go build -o gs-server cmd/gs-server/main.go
```

### 3. Run the GS Server

Run the `gs-server` on your Ground Station, pointing it to the Air Unit's IP address:

```bash
./gs-server -airunit http://192.168.1.10:8080 -listen :8081 -static ./web/dist
```

- `-airunit`: URL of the Air Unit API (default: `http://192.168.1.10:8080`).
- `-listen`: Address to listen on (default: `:8081`).
- `-static`: Path to the compiled frontend files (default: `./web/dist`).
- `-config`: Path to the local `wifibroadcast.cfg` file (default: `/etc/wifibroadcast.cfg`). This is used to update local radio settings when changed via the WebUI.

Access the WebUI in your browser at `http://localhost:8081`.

## API Endpoints

### Radio (`/api/v1/radio`)
*Manages WFB-ng wireless settings.*

- **GET**: Retrieve current settings.
- **POST**: Update settings.
  ```bash
  curl -X POST -d '{"channel":161, "bandwidth":20, "tx_power":55}' http://localhost:8080/api/v1/radio
  ```

### Video (`/api/v1/video`)
*Manages Majestic video encoding.*

- **GET**: Retrieve video settings.
- **POST**: Update settings.
  ```bash
  curl -X POST -d '{"resolution":"1920x1080", "fps":60, "bitrate":7000}' http://localhost:8080/api/v1/video
  ```

### Camera (`/api/v1/camera`)
*Manages Camera sensor settings.*

- **GET**: Retrieve camera settings.
- **POST**: Update settings.
  ```bash
  curl -X POST -d '{"flip":true, "mirror":false, "contrast":50}' http://localhost:8080/api/v1/camera
  ```

### Adaptive Link (`/api/v1/adaptive-link`)
*Manages Adaptive Link logic.*

- **GET**: Retrieve settings.
- **POST**: Update settings. Toggling `enabled` updates `/etc/rc.local` to start/stop the `alink_drone` process on boot.
  ```bash
  curl -X POST -d '{"enabled":true, "allow_set_power":true, "power_level_0_to_4":3}' http://localhost:8080/api/v1/adaptive-link
  ```

## Development / Testing

You can run the service locally by setting environment variables to override the default configuration paths:

```bash
export WFB_PATH=./test_configs/wfb.yaml
export MAJESTIC_PATH=./test_configs/majestic.yaml
export ALINK_PATH=./test_configs/alink.conf
export RC_LOCAL_PATH=./test_configs/rc.local
export INIT_D_PATH=./test_configs/init.d

go run cmd/api/main.go
```
