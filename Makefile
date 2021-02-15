ALL_OS = linux
ALL_ARCH = amd64 arm64


define BUILD_TARGET =
docker-name-resolver-$(1)-$(2): main.go
	CGO_ENABLED=0 GOOS=$(1) GOARCH=$(2) go build -o $$@ -ldflags="-s -w" $$<
endef

.PHONY: all

all: $(foreach os,$(ALL_OS), $(foreach arch, $(ALL_ARCH), docker-name-resolver-$(os)-$(arch)))

$(foreach os,$(ALL_OS), $(foreach arch, $(ALL_ARCH), $(eval $(call BUILD_TARGET,$(os),$(arch)))))
