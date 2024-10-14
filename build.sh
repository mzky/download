#!/bin/bash
app="download"
go fmt ./...;goimports -w .
go build -ldflags "-w -s" -o "${app}" main.go
upx "${app}"
