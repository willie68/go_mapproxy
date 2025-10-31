# go_mapproxy

## Preface

**go_mapproxy** is a lightweight, high-performance proxy application for Slippy Map (XYZ), Tile Map Services (TMS), WMS services and MBTiles files, written in Go. It is designed to provide fast and efficient access to map tiles from various sources, with optional caching and prefetching capabilities. The application is suitable for both production use and for developers who want to build or optimize their own mapping solutions. 

The original primary purpose was to provide a lightweight service that could be installed on your map client and would proxy XYZ requests to WMS. (see https://github.com/willie68/MCSDepthLoggerUI)

------

## Intention

The main goals of **go_mapproxy** are:

- **Proxy Functionality:** Forward XYZ (Slippy map) requests to a XYZ, TMS or WMS map server. Or use a MBTiles File as tile source.
- **Caching:** Optionally cache tiles to speed up access and reduce server load.
- **Prefetching:** Preload tiles for defined zoom levels and map providers.
- **Simple Configuration:** Use a clear YAML configuration file for quick and easy setup.

------

## Installation

### with GO

There are many ways to install this little proxy. 

First if you have already golang installed, simply do an

`go install github.com/willie68/go_mapproxy@latest`

After that you can simply start the proxy with `go_mapproxy`

### with Download

There are prebuilded binaries availble in the github repo under 

https://github.com/willie68/go_mapproxy/releases

Take the binary for your OS and extract it into a folder of your choise. There you can start the gomapproxy binary. (don't forget the set the executable flag)

### with Docker

another way is to start this as a docker container. As there are no prebuild images, you have to download and extract the repository. After that start

`docker build -f ./dockerfile -t gomapproxy:latest ./` 

and then start the image with 

`docker run --restart=always --name gomapproxy -p 8580:8580 gomapproxy:latest`

#### providing a config with docker

If you like to provide another config for the map proxy, you can put the config into a directory on your host. e.g. i'll use e:\daten\docker\gomapproxy as a root for all gomapproxy on docker related files. The structure here ist something like that:

```
e:\daten\docker\gomapproxy\
   +-- cache\ <- folder for the cache
   +-- config.yaml <- the configuration for the service
   +-- OSM-OpenCPN2-Adria.mbtiles <- a mbtiles file of the adria region from OpenSeaMap
   
```

The config for that is here (for a description see [Configuration Examples](#Configuration Examples):

```yaml
http:
  port: 8580
  sslport: 8443 # set to e.g. 8443 to enable https server

cache:
  active: true
  path: /opt/gomapproxy/cache
  maxage: 168 # in hours, 7d * 24h = 168h

logging:
  level: debug

provider: 
  gebco:
    url: https://geoserver.openseamap.org/geoserver/gwc/service/wms
    type: wms
    layers: gebco2021:gebco_2021
    format: image/png
    nocache: false
  osmde:
    url:  https://tile.openstreetmap.de
    type: xyz
    format: image/png
    nocache: false
    noprefetch: true
    headers:
      Accept: image/png
      User-Agent: gomapproxy v0.1.8
  adria:
    type: mbtiles
    path: /opt/gomapproxy/mbtiles/OSM-OpenCPN2-Adria.mbtiles
    nocache: true
    fallback: osmde

```

start this service with
`docker run -d --restart=always --name gomapproxy -p 8580:8580 -p 8443:8443 -v e:\daten\docker\gomapproxy\config.yaml:/config/config.yaml -v e:\daten\docker\gomapproxy\cache:/opt/gomapproxy/cache -v e:\daten\docker\gomapproxy\OSM-OpenCPN2-Adria.mbtiles:/opt/gomapproxy/mbtiles/OSM-OpenCPN2-Adria.mbtiles gomapproxy:latest`

- `-d`: just start the container and come back to terminal
- `-restart=always`  will always restart the service after reboot
- `--name gomapproxy`  name of the running container
- `-p...` map the internal ports 8580 and 8443 to the same external ports. If you need other ports on your host, change the first parameter `-p 8080:8580` will use the 8080 port on the host for the http interface. (If you set a ssl port, only that port will provide the tiles)
- `-v`: here comes the mapping part:
  - first map the config.yaml file of the host to the internal location
  - second map is for the cache directory. After start you will find all the cache files in your host directory
  - third mapping is for the mbtiles file
- the last part is the name of the image

After that you can see in your container with the `docker ps` command. There should be a line like this:

```
CONTAINER ID   IMAGE                 COMMAND                  CREATED         STATUS          PORTS                                                                                      NAMES
cccf0c03e563   gomapproxy:latest     "/service --config /…"   3 minutes ago   Up 3 minutes    0.0.0.0:8580->8580/tcp, [::]:8580->8580/tcp, 0.0.0.0:9443->9443/tcp, [::]:9443->9443/tcp   gomapproxy
```



## Usage

### Basic Proxy

Start the proxy with a configuration file:

`gomapproxy -c config.yaml`

### Proxy with Caching

1. Enable caching and set a cache path in your configuration.

   ```yaml
   caching:
     active: true
     path: ./tilecache
     maxage: 168 # in hours, 7d * 24h = 168h
   ```

2. Start the application:

`gomapproxy -c config.yaml`

### Proxy with Prefetching

To prefetch tiles up to a certain zoom level for a specific provider:

`gomapproxy -c config.yaml -s <providernames as csv> -z 4`

### Check functionality
if you want to try, that your proxy is working simply load a tile. The URL for such a request is 
`http://[your hostname]:[port]/tileserver/[provider]/[z]/[x]/[y].png` 

e.g. `http://localhost:8580/tileserver/osm/xyz/4/8/5.png`

### Command Line Options

- `-c, --config`: Path to the configuration file (default: config.yaml)
- `-p, --port`: Overwrite the port specified in the config
- `-i, --init`: Write out a default config file
- `-v, --version`: Show the current version
- `-z, --zoom`: Max zoom for prefetch tiles
- `-s, --system`: Prefetch provider (comma-separated for multiple provider)

------

## Configuration Examples

### Minimal Example without caching

```yaml
http:
  port: 8580
  sslport: 0 # set to e.g. 8443 to enable https server

caching:
  active: false
  path: ./tilecache
  maxage: 168 # in hours, 7d * 24h = 168h

provider:
  gebco:
    url: https://geoserver.openseamap.org/geoserver/gwc/service/wms
    type: wms
    layers: gebco2021:gebco_2021
    format: image/png
    cached: false

```

### Prefetch for Multiple Systems

`gomapproxy -c config.yaml -s "gebco,osm" -z 5`

------

## Further Information

- Full documentation and examples:
  [Project Readme](https://github.com/willie68/go_mapproxy)
- For questions or support, please open an issue on GitHub.

------

**Note:**
This application is cross-platform and can be run on any system with a Go runtime.

## internal workflow: How this will work
- take a xyz request
- check provider
- if provider is cached, try cache -> ok, return tile
- -> not ok, check provider type 
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
  - mbtiles
    - check if the zoom level is ok, not ok -> empty.png or fallback
    - check the bounding box of the mbtiles metadata: not ok -> empty.png or fallback
    - try to read the tile from the MBTiles file: not ok -> empty.png or fallback
  - if configured and provider is cacheable, cache the tile

## Restrictions
- only 256x256px tiles possible
- only srs=EPSG:3857 is possible
- no server description is proxied

## Caching

For performance boost you can enable caching of tiles. In the config enable the cache. 

```yaml
caching:
  active: true
  path: ./tilecache
  maxage: 168 # in hours, 7d * 24h = 168h
```

`active`: set to true to activate tile caching

`path`: path where the `gomapproxy` can store tiles. (Keep in mind how much storage you may need.)
`maxage`: setting the maximal age of tiles in hours. If a tile is older, the background process will automatically delete this tile and the app logic will no loger distribute this tile. 

Second [optional]: if a provider should not be cached, use the nocache option

```yaml
provider:
 osm:
  url:  https://tile.openstreetmap.org
  type: xyz
  format: image/png
  nocache: true
```

`nocache`: set this property to true and all tiles requested from this server will not be stored into the cache.

The cache will store the tiles file by a double subfolder structure based on the file hash. And than the tile metadata will be stored in a key/value store database, key is the metadata (provider, x,y,z), value the hash of the tile. As the hash is unique for the tiles, tiles with identically content will have the same hash. And will be stored only once. (e.g. like tiles of the ocean) The database is stored in the subdirectory `badger` (as it's a badgerdb) and the tiles will be stored in a sub folder `tiles`. (Single-Instance-Storage)

## Provider configuration

```yaml
provider:
  <provider name>:
    url:  https://tile.openstreetmap.org
    type: xyz
    layers: # only for wms servers
    format: image/png
    version: 1.1.0 # only for wms servers
    nocache: false
    noprefetch: false
    path: # path to the mbtiles file, for mbtiles only
    styles: # only for wms servers
    fallback: <provider name> # fallback provider
    headers:
     Accept: image/png,image/jpg,*/*;q=0.8
     User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:143.0) Gecko/20100101 Firefox/143.0
```

`url`: the url of the original tile server
`type`: the type of server, xyz, tms or wms
`layers` : only used for the wms type. The layer of the wms to be used
`format`: the format of the tiles, returned by the server. No format conversion will be done.
`version` : the version of the responses of the wms server. Only 1.1.0 and 1.3.0 are supported
`nocache`: true to deactivate caching of this provider 
`path` : path to the mbtiles file, for mbtiles provider only
`noprefetch` : this provioder will not allow prefetching. There are some provider, who doesn't allow prefetching, like the osm. If you want to prevent prefetching, set this option to true. (There is an internal blacklist, too) 
`styles` : some style setting for wms servers 
`fallback` : for mbtiles you can set here an fallback provider. If a tile is not served from the mbtiles file, the app will try to read the file from this provider. Otherwise an empty.png will be displayed.
`header`: add additional headers, as they may be needed by the provided tile server (like osm)

## Setting up TLS

There are two ways to set up this service with tls, depending if you want to use an already create certificate ( Let's Encrypt as example) or you're ok using self signed certificates.
For the latter simply add an valid port for the `sslport` config key. A new certificate will be generated automatically on every start.

```yaml
http:
  port: 8580
  sslport: 8443 # set to e.g. 8443 to enable https server
  certificate: # path and name to the certificate (PEM Format)
  key: # path and name to the private key (PEM Format)
```

IF you use an already created certificate simply add the path and name of the PEM files to the config parameters certificate (public part) and key (private part).

## A word on prefetching of tiles

You can prefetch single/multiple provider with the `system` and `zoom` parameter. All tiles of the the selected provider from 0 to zoom will be prefetched. (At this time no prefetch bonding boxes are configurable) Be aware you need the space for that. Prefechting with level 8 is round about 1GB. (depends on the wms provider) Level 9 ~ 5GB... (And it will take some time)

example: `gomapproxy -c config.yaml -s gebco -z 9`

This will prefetch all tiles from the server with the alias gebco for zoom levels 0 to 9.
But be aware, some providers as the osm don't allow prefetching. You can swithc prefechting of in the config, but for some providers (like openstreetmap) will be automatically ignored on prefetch. 
