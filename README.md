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
GOOS=linux GOARCH=arm go build -o ezconfig cmd/api/main.go
```

## Installation & Usage

1.  Copy the `ezconfig` binary to the Air Unit (e.g., via SCP).
2.  Make it executable: `chmod +x ezconfig`.
3.  Run the service: `./ezconfig`.

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
