APP_NAME := PassIt
ICON     := assets/icon.png
APP_ID   := com.p4tin.passit
UNAME    := $(shell uname)

.PHONY: build run _run-macos _run-other test vet \
        package-darwin package-windows package-linux \
        cross-darwin cross-windows cross-linux cross-all \
        install-tools clean

# ── Local dev ─────────────────────────────────────────────────────────────────

build:
	go build -o $(APP_NAME) ./cmd/passit

# On macOS: wraps the binary in a minimal .app bundle so the Dock shows the
# correct icon and name. Uses sips (macOS built-in) to convert the PNG to ICNS.
# On Linux/Windows: runs the binary directly.
run: build
ifeq ($(UNAME), Darwin)
	@$(MAKE) --no-print-directory _run-macos
else
	@$(MAKE) --no-print-directory _run-other
endif

_run-macos:
	@rm -rf $(APP_NAME).app /tmp/passit.iconset
	@mkdir -p $(APP_NAME).app/Contents/MacOS $(APP_NAME).app/Contents/Resources
	@cp $(APP_NAME) $(APP_NAME).app/Contents/MacOS/$(APP_NAME)
	@mkdir -p /tmp/passit.iconset
	@cp $(ICON) /tmp/passit.iconset/icon_32x32@2x.png
	@iconutil -c icns /tmp/passit.iconset -o $(APP_NAME).app/Contents/Resources/icon.icns
	@rm -rf /tmp/passit.iconset
	@cp scripts/Info.plist $(APP_NAME).app/Contents/Info.plist
	open -W $(APP_NAME).app

_run-other:
	./$(APP_NAME)

test:
	go test ./...

vet:
	go vet ./...

# ── Native packaging (run on the target OS) ───────────────────────────────────
# Requires: make install-tools

package-darwin:
	cd cmd/passit && fyne package -os darwin -icon ../../$(ICON) -name $(APP_NAME) -appID $(APP_ID)
	@mv cmd/passit/$(APP_NAME).app . 2>/dev/null || true
	@echo "→ $(APP_NAME).app"

package-windows:
	cd cmd/passit && fyne package -os windows -icon ../../$(ICON) -name $(APP_NAME) -appID $(APP_ID)
	@mv cmd/passit/$(APP_NAME).exe . 2>/dev/null || true
	@echo "→ $(APP_NAME).exe"

package-linux:
	cd cmd/passit && fyne package -os linux -icon ../../$(ICON) -name $(APP_NAME) -appID $(APP_ID)
	@mv cmd/passit/$(APP_NAME) ./$(APP_NAME)-linux 2>/dev/null || true
	@echo "→ $(APP_NAME)-linux"

# ── Cross-platform packaging via fyne-cross (requires Docker) ─────────────────
# Requires: make install-tools  +  Docker running

cross-darwin:
	fyne-cross darwin -arch=amd64,arm64 -icon $(ICON) -name $(APP_NAME) -app-id $(APP_ID) ./cmd/passit

cross-windows:
	fyne-cross windows -arch=amd64 -icon $(ICON) -name $(APP_NAME) -app-id $(APP_ID) ./cmd/passit

cross-linux:
	fyne-cross linux -arch=amd64,arm64 -icon $(ICON) -name $(APP_NAME) -app-id $(APP_ID) ./cmd/passit

cross-all: cross-darwin cross-windows cross-linux

# ── Tooling ───────────────────────────────────────────────────────────────────

install-tools:
	go install fyne.io/fyne/v2/cmd/fyne@latest
	go install github.com/fyne-io/fyne-cross@latest

clean:
	rm -rf $(APP_NAME) $(APP_NAME).app $(APP_NAME).exe $(APP_NAME)-linux fyne-cross/
