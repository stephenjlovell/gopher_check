#!/bin/bash

# This script partly automates creation of gopher_check binaries for new releases.

createBinaries() {
  read -p "enter the current semantic version number: " version
  echo "creating binaries for version: ${version}"
  mkdir "./binaries"
  versionPath="./binaries/${version}"
  mkdir "${versionPath}"

  mkdir "${versionPath}/mac"
  env GOARCH=amd64 GOOS=darwin go build -o "${versionPath}/mac/gopher_check"
  zip "${versionPath}/mac/gopher_check-${version}-mac-amd64.zip" "${versionPath}/mac/gopher_check"

  mkdir "${versionPath}/windows"
  env GOARCH=amd64 GOOS=windows go build -o "${versionPath}/windows/gopher_check.exe"
  zip "${versionPath}/windows/gopher_check-${version}-windows-amd64.zip" "${versionPath}/windows/gopher_check.exe"

  mkdir "${versionPath}/linux"
  env GOARCH=amd64 GOOS=linux go build -o "${versionPath}/linux/gopher_check"
  zip "${versionPath}/linux/gopher_check-${version}-linux-amd64.zip" "${versionPath}/linux/gopher_check"
}

createBinaries
