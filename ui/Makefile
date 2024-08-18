build:
	GOARCH=wasm GOOS=js go build -o web/app.wasm
	go build -o web/b2wui-server

run: build
	./web/b2wui-server