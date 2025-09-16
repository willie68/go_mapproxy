@echo off
echo building go_mapproxy
rem goreleaser build --snapshot --clean --single-target
rem cd dist\osml_windows_amd64_v1
rem osml.exe version
go build -ldflags="-s -w" -o ./dist/gomapproxy.exe cmd/main.go
copy .\dist\gomapproxy.exe c:\tools\
rem cd ..\..