#!/bin/bash
app=download
statik -src=./static
go fmt ./...;goimports -w .
go build -ldflags "-w -s" -o "${app}" main.go
upx "${app}"
