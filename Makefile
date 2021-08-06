VERSION=`git tag | head -1`
VERSION_LATEST=`git tag | tail -n 1`
VERSION_MASTER=master
COMMIT=`git rev-parse --short HEAD`
BUILDDATE=`date "+%Y-%m-%d/%H:%M:%S"`
BUILD_DIR=build
APP_NAME=dtm
sources=$(wildcard *.go)

build = GOOS=$(1) GOARCH=$(2) go build -o ${BUILD_DIR}/$(APP_NAME)-$(1)-$(2) -ldflags "-w -s -X main.a=${APP_NAME} -X main.v=${VERSION} -X main.c=${COMMIT} -X main.d=${BUILDDATE}" main.go
md5 = md5sum ${BUILD_DIR}/$(APP_NAME)-$(1)-$(2) > ${BUILD_DIR}/$(APP_NAME)-$(1)-$(2)_checksum.txt
tar =  tar -cvzf ${BUILD_DIR}/$(APP_NAME)-$(1)-$(2).tar.gz  -C ${BUILD_DIR}  $(APP_NAME)-$(1)-$(2) $(APP_NAME)-$(1)-$(2)_checksum.txt
delete = rm -rf ${BUILD_DIR}/$(APP_NAME)-$(1)-$(2) ${BUILD_DIR}/$(APP_NAME)-$(1)-$(2)_checksum.txt

mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))

CURRENT_DIR := $(notdir $(patsubst %/,%,$(dir $(mkfile_path))))

ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

ALL_LINUX = linux-amd64 \
	linux-386 \
	linux-arm \
	linux-arm64

ALL = $(ALL_LINUX) \
		darwin-amd64

build_linux: $(ALL_LINUX:%=build/%)

build_all: $(ALL:%=build/%)

build/%:
	$(call build,$(firstword $(subst -, , $*)),$(word 2, $(subst -, ,$*)))
	$(call md5,$(firstword $(subst -, , $*)),$(word 2, $(subst -, ,$*)))
	$(call tar,$(firstword $(subst -, , $*)),$(word 2, $(subst -, ,$*)))
	$(call delete,$(firstword $(subst -, , $*)),$(word 2, $(subst -, ,$*)))

clean:
	rm -rf ${BUILD_DIR}
build:
	go build -v -ldflags "-w -s -X main.a=${APP_NAME} -X main.v=${VERSION_MASTER} -X main.c=${COMMIT} -X main.d=${BUILDDATE}" -o ${BUILD_DIR}/${APP_NAME} main.go

vet:
	go vet main.go

linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${BUILD_DIR}/${APP_NAME} -ldflags "-w -s -X main.a=${APP_NAME} -X main.v=${VERSION_MASTER} -X main.c=${COMMIT} -X main.d=${BUILDDATE}" main.go