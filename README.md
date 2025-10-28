# go_mapproxy

## Preface

**go_mapproxy** is a lightweight, high-performance proxy application for Slippy Map (xyz), Tile Map Services (TMS), WMS services and MBTiles files, written in Go. It is designed to provide fast and efficient access to map tiles from various sources, with optional caching and prefetching capabilities. The application is suitable for both production use and for developers who want to build or optimize their own mapping solutions. 

The original primary purpose was to provide a lightweight service that could be installed on your map client and would proxy XYZ requests to WMS. (see https://github.com/willie68/MCSDepthLoggerUI)

------

## Intention

The main goals of **go_mapproxy** are:

- **Proxy Functionality:** Forward XYZ requests to a XYZ, TMS or WMS map server. Or use a MBTiles File.
- **Caching:** Optionally cache tiles to speed up access and reduce server load.
- **Prefetching:** Preload tiles for defined zoom levels and map providers.
- **Simple Configuration:** Use a clear YAML configuration file for quick and easy setup.

------

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
`http://[your hostname]:[port]/[provider]/[z]/[x]/[y].png` 

e.g. `http://localhost:8580/osm/xyz/4/8/5.png`

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
port: 8580
caching:
  active: false
  path: ./tilecache
  maxage: 2160 # in hours, 90 days = 2160

provider:
  gebco:
    url: https://geoserver.openseamap.org/geoserver/gwc/service/wms
    type: wmss
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
  osm:
    url:  https://tile.openstreetmap.org
    type: xyz
    layers: # only for wms servers
    format: image/png
    version: 1.1.0 # only for wms servers
    nocache: false
    noprefetch: false
    styles: # only for wms servers
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

## A word on prefetching of tiles

You can prefetch single/multiple provider with the `system` and `zoom` parameter. All tiles of the the selected provider from 0 to zoom will be prefetched. (At this time no prefetch bonding boxes are configurable) Be aware you need the space for that. Prefechting with level 8 is round about 1GB. (depends on the wms provider) Level 9 ~ 5GB... (And it will take some time)

example: `gomapproxy -c config.yaml -s gebco -z 9`

This will prefetch all tiles from the server with the alias gebco for zoom levels 0 to 9.
