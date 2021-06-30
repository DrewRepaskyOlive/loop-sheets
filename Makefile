
clean:
	rm -rf ./build

build/darwin-amd64/olive-helps-loop:
	GOOS=darwin GOARCH=amd64 go build  \
	     -o build/darwin-amd64/olive-helps-loop

build/windows-amd64/olive-helps-loop:
	GOOS=windows GOARCH=amd64 go build  \
	     -o build/windows-amd64/olive-helps-loop.exe

build: clean build/darwin-amd64/olive-helps-loop build/windows-amd64/olive-helps-loop

.PHONY: clean build build/darwin-amd64/olive-helps-loop build/windows-amd64/olive-helps-loop
