#!/bin/bash
GOOS=darwin GOARCH=386 go build -o cloudupload_darwin_386
GOOS=darwin GOARCH=amd64 go build -o cloudupload_darwin_amd64
GOOS=freebsd GOARCH=386 go build -o cloudupload_freebsd_386
GOOS=freebsd GOARCH=amd64 go build -o cloudupload_freebsd_amd64
GOOS=linux GOARCH=386 go build -o cloudupload_linux_386
GOOS=linux GOARCH=amd64 go build -o cloudupload_linux_amd64
GOOS=linux GOARCH=arm64 go build -o cloudupload_linux_arm64
GOOS=netbsd GOARCH=386 go build -o cloudupload_netbsd_386
GOOS=netbsd GOARCH=amd64 go build -o cloudupload_netbsd_amd64
GOOS=openbsd GOARCH=386 go build -o cloudupload_openbsd_386
GOOS=openbsd GOARCH=amd64 go build -o cloudupload_openbsd_amd64
GOOS=openbsd GOARCH=arm64 go build -o cloudupload_openbsd_arm64

rm cloudupload.tgz
tar czf cloudupload.tgz cloudupload_*
rm -rf cloudupload_*
