#!/bin/bash

# Ubuntu 20.04 LTS
GOOS=linux 
GOARCH=amd64

go build -o bin/httpget httpget.go
go build -o bin/httppost httppost.go
go build -o bin/httpdownload httpdownload.go
go build -o bin/httpupload httpupload.go