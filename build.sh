#!/bin/bash
git describe >git-describe.txt
go build -o installer-linux-amd64

