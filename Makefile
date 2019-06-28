NAMESPACE := github.com/DCSO/mauerspecht
CLIENT_PLATFORMS := x86_64-linux x86_64-windows x86_64-darwin i386-linux i386-windows
GOFILES := $(shell find -name "*.go")

define goarch
$(or $(if $(findstring x86_64,$1),amd64),\
     $(if $(findstring i386,$1),386),\
     $(error unknown arch $1))
endef

define goos
$(or $(if $(findstring linux,$1),linux),\
     $(if $(findstring windows,$1),windows),\
     $(if $(findstring darwin,$1),darwin),\
     $(error unknown os $1))
endef

define extension
$(if $(findstring windows,$1),.exe)
endef

CLIENT_BINARIES := $(foreach platform,$(CLIENT_PLATFORMS),\
	mauerspecht-client.$(platform)$(call extension,$(platform)))

$(foreach platform,$(CLIENT_PLATFORMS),\
	$(eval %.$(platform)$(call extension,$(platform)): export GOOS=$(call goos,$(platform)))\
	$(eval %.$(platform)$(call extension,$(platform)): export GOARCH=$(call goarch,$(platform)))\
)

.PHONY: all
all: mauerspecht-server mauerspecht-clients

mauerspecht-server:
	go build -o $@ $(NAMESPACE)/cmd/server

.PHONY: mauerspecht-clients
mauerspecht-clients: $(CLIENT_BINARIES)

mauerspecht-client.%: private subdir=cmd/client
mauerspecht-client.%: $(GOFILES)
	go build -o $@ $(NAMESPACE)/$(subdir)

.PHONY: clean
clean:
	rm -f mauerspecht-server $(CLIENT_BINARIES)
