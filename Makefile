# Copyright 2017 Orange
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Basic settings
GO = go
BINARY_NAME = custom_exporter.exe
CONFIG_FILE = example.yml

# Build flags
BUILDFLAGS = -tags windows
LDFLAGS = -s -w

# Environment variables
export CGO_ENABLED=0
export GOOS=windows
export GOARCH=amd64

# Targets
.PHONY: all clean format test vet build run

all: format vet test build

clean:
	@echo Cleaning...
	@$(GO) clean
	@if exist $(BINARY_NAME) del /f $(BINARY_NAME)

format:
	@echo Formatting code...
	@$(GO) fmt ./...

test:
	@echo Running tests...
	@$(GO) test $(BUILDFLAGS) ./...

vet:
	@echo Vetting code...
	@$(GO) vet $(BUILDFLAGS) ./...

build:
	@echo Building...
	@$(GO) build -v $(BUILDFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME)

run: build
	@echo Running with example config...
	@.\$(BINARY_NAME) -collector.config $(CONFIG_FILE)
