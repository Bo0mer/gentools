#!/bin/bash

set -e

go install
go generate ./...
ginkgo -r --randomizeAllSpecs
