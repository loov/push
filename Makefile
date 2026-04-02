.PHONY: all
all: fmt test vet staticcheck

# --- formatting ---

.PHONY: fmt
fmt:
	goimports -w -local github.com/loov .

# --- testing ---

.PHONY: test
test:
	go test -race ./...

# --- vetting and staticcheck for common platforms ---

#PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64
PLATFORMS := darwin/arm64

VET_TARGETS := $(addprefix vet/,$(PLATFORMS))
STATICCHECK_TARGETS := $(addprefix staticcheck/,$(PLATFORMS))

.PHONY: vet $(VET_TARGETS)
vet: $(VET_TARGETS)
$(VET_TARGETS): vet/%:
	GOOS=$(word 1,$(subst /, ,$*)) GOARCH=$(word 2,$(subst /, ,$*)) go vet ./...

.PHONY: staticcheck $(STATICCHECK_TARGETS)
staticcheck: $(STATICCHECK_TARGETS)
$(STATICCHECK_TARGETS): staticcheck/%:
	GOOS=$(word 1,$(subst /, ,$*)) GOARCH=$(word 2,$(subst /, ,$*)) staticcheck ./...

.PHONY: quickcheck
quickcheck:
	go vet ./...
	staticcheck ./...

# --- MIDI Device Script install ---

MIDI_DEVICE_SCRIPT_SRC := unPush 3.device
MIDI_DEVICE_SCRIPT_DST := $(HOME)/Music/Audio Music Apps/MIDI Device Scripts/Ableton/unPush 3.device

.PHONY: install-midi-device-script
install-midi-device-script:
	@mkdir -p "$(MIDI_DEVICE_SCRIPT_DST)"
	cp -R "$(MIDI_DEVICE_SCRIPT_SRC)/" "$(MIDI_DEVICE_SCRIPT_DST)/"
	@echo 'Installed to "$(MIDI_DEVICE_SCRIPT_DST)"'
	@echo 'Restart Logic Pro to pick up changes.'
