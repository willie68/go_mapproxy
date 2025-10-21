# go_mapproxy

## Preface

**go_mapproxy** is a lightweight, high-performance proxy application for Slippy Map, Tile Map Services (TMS) and WMS services, written in Go. It is designed to provide fast and efficient access to map tiles from various sources, with optional caching and prefetching capabilities. The application is suitable for both production use and for developers who want to build or optimize their own mapping solutions. 

The original primary purpose was to provide a lightweight service that could be installed on your map client and would proxy XYZ requests to WMS. (see https://github.com/willie68/MCSDepthLoggerUI)

------

## Intention

The main goals of **go_mapproxy** are:

- **Proxy Functionality:** Forward XYZ requests to a XYZ, TMS or WMS map server.
- **Caching:** Optionally cache tiles to speed up access and reduce server load.
- **Prefetching:** Preload tiles for defined zoom levels and map systems.
- **Simple Configuration:** Use a clear YAML configuration file for quick and easy setup.

------

## Usage

### Basic Proxy

Start the proxy with a configuration file:

`gomapproxy -c config.yaml`

### Proxy with Caching

1. Enable caching and set a cache path in your configuration.

2. Start the application:

`gomapproxy -c config.yaml`

### Proxy with Prefetching

To prefetch tiles up to a certain zoom level for a specific system:

`gomapproxy -c config.yaml -s <systemname> -z 4`

### Command Line Options

- `-c, --config`: Path to the configuration file (default: [config.yaml](vscode-file://vscode-app/c:/Users/wklaa/AppData/Local/Programs/Microsoft VS Code/resources/app/out/vs/code/electron-browser/workbench/workbench.html))
- `-p, --port`: Overwrite the port specified in the config
- `-i, --init`: Write out a default config file
- `-v, --version`: Show the current version
- `-z, --zoom`: Max zoom for prefetch tiles
- `-s, --system`: Prefetch system (comma-separated for multiple systems)

------

## Configuration Examples

### Minimal Example 

```yaml
port: 8580
caching:
  active: true
  path: ./tilecache
  maxage: 2160 # in hours, 90 days = 2160

tileservers:
  gebco:
    url: https://geoserver.openseamap.org/geoserver/gwc/service/wms
    type: wmss
    layers: gebco2021:gebco_2021
    format: image/png
    cached: true

```



### Prefetch for Multiple Systems

`gomapproxy -c config.yaml -s "gebco,osm" -z 5`

------

## Further Information

- Full documentation and examples:
  [https://github.com/willie68/go_mapproxy](vscode-file://vscode-app/c:/Users/wklaa/AppData/Local/Programs/Microsoft VS Code/resources/app/out/vs/code/electron-browser/workbench/workbench.html)
- For questions or support, please open an issue on GitHub.

------

**Note:**
This application is cross-platform and can be run on any system with a Go runtime.

## How this will work
- take a xyz request
- check system
- if system cached, try cache -> ok, return tile
- -> not ok, check system type 
  - wms
    - convert xyz to wms bounding box
    - do the wms request on the desired server, server configurable in a config 
    - proxy the answered png to the requesting client
  - tms
    - invert y coordinate
    - do the tms request on the desired server, server configurable in a config 
    - proxy the answered png to the requesting client
  - xyz
    - do the tms request on the desired server, server configurable in a config 
    - proxy the answered png to the requesting client
  - if configured and system cachable, cache the tile

## Restrictions
- only 256x256px tiles possible
- only srs=EPSG:3857 is possible
- no server description is proxied

## A word on prefetching of tiles

You can prefetch single/multiple system with the `system` and `zoom` parameter. All tiles of the the systems from 0 to zoom will be prefetched. (At this time no prefetch bonding boxes are configurable) Be aware you need the space for that. Prefechting with level 8 is round about 1GB. (depends on the wms provider) Level 9 ~ 5GB... (And it will take some time)

example: `gomapproxy -c config.yaml -s gebco -z 9`

This will prefetch all tiles from the server with the alias gebco for zoom levels 0 to 9.
