#!/bin/bash
set -x

set CGO_ENABLED=0 
go test ./... -coverprofile="ut.cover" -covermode count -v -json . 2>&1 | tee "test_report.log" || true) && \
go-junit-report -parser gojson -in "test_report.log" -out "report.xml" && \
gocover-cobertura < ut.cover > coverage.xml && \
gosec -no-fail -fmt=sonarqube -out=gosec.json ./...

