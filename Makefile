WAILS := $(shell go env GOPATH)/bin/wails
APP   := build/bin/TUIStudio.app
DMG   := build/bin/TUIStudio.dmg

.PHONY: dev build package clean studio

## Start Vite dev server + Wails window (hot reload)
dev:
	cd web && npm run dev &
	$(WAILS) dev --frontenddevserverurl http://localhost:5173

## Compile the .app bundle
build:
	$(WAILS) build

## Compile the .app bundle and package into a .dmg
package: build
	@rm -f $(DMG)
	create-dmg \
		--volname "TUIStudio" \
		--window-pos 200 120 \
		--window-size 660 400 \
		--icon-size 128 \
		--icon "TUIStudio.app" 180 170 \
		--hide-extension "TUIStudio.app" \
		--app-drop-link 480 170 \
		$(DMG) \
		$(APP)
	@echo "→ $(DMG)"

## Remove build output
clean:
	rm -rf build/bin

## Interactive action picker (requires huh)
studio:
	go run ./cmd/studio
