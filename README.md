# go_mapproxy
A minimal mapproxy for converting TMS request to an WMS Tileserver, written in go, using slippy map ordered tiles.
https://wiki.openstreetmap.org/wiki/Slippy_map_tilenames

## features
- take a tms request
- convert xyz to wms bounding box
- do the wms request on the desired server, server configurable in a config 
- proxy the answered png to the requesting client
- if configured, a simple file cache is applied

## restrictions
- only 256x256px tiles possible
- only srs=EPSG:3857

## configuration
see configs/config.yaml


