# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#
# Usage:
# 	[PREFIX=gsosa/redirect] [ARCH=amd64] [TAG=1.1.0] make (server|container|push)


all: push

TAG=1.0
PREFIX?=gsosa/redirect
ARCH?=amd64
GOLANG_VERSION=1.7
TEMP_DIR:=$(shell mktemp -d)

server: server.go
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) GOARM=6 go build -a -installsuffix cgo -ldflags '-w -s' -o server

container:
	# Compile the binary inside a container for reliable builds
	docker pull golang:$(GOLANG_VERSION)
	docker run --rm -it -v $(PWD):/go/src/redirect golang:$(GOLANG_VERSION) /bin/bash -c "make -C /go/src/redirect server ARCH=$(ARCH)"

build:
	docker build --pull -t $(PREFIX):$(TAG) .

push: server build
	docker login -u="$(DOCKER_USERNAME)" -p="$(DOCKER_PASSWORD)"
	docker push $(PREFIX):$(TAG)

clean:
	rm -f server
