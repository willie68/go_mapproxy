# go_mapproxy
A minimal mapproxy for converting TMS request to an WMS Tileserver, written in go, using slippy map ordered tiles.
https://wiki.openstreetmap.org/wiki/Slippy_map_tilenames

## work steps
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

## prefetching of tiles

You can prefetch a single system with the `system` and `zoom` parameter. All tiles of the the system from 0 to zoom will be prefetched. Be aware you need the space for that. Prefechting with level 8 is round about 1GB. (depends on the wms provider) Level 9 ~ 5GB... (And it will take some time)
example: `gomapproxy .c config.yaml -s gebco -z 9`

This will prefetch all tiles from the server with the alias gebco from zomm level 0 to 9.
