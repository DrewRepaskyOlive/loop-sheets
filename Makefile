
clean:
	rm -rf ./build

build/darwin-amd64/plugin:
	GOOS=darwin GOARCH=amd64 go build  \
	     -o build/darwin-amd64/plugin

build/windows-amd64/plugin:
	GOOS=windows GOARCH=amd64 go build  \
	     -o build/windows-amd64/plugin.exe

build: clean build/darwin-amd64/plugin build/windows-amd64/plugin
