@echo off
echo building go_mapproxy
goreleaser build --snapshot --clean --single-target
cd dist\gomapproxy_windows_amd64_v1
gomapproxy.exe --version
copy gomapproxy.exe c:\tools\
rem go build -ldflags="-s -w" -o ./dist/gomapproxy.exe cmd/main.go
rem copy .\dist\gomapproxy.exe c:\tools\
cd ..\..