windows:
	env GOOS=windows GOARCH=amd64 go build -o kustomize-graph-windows-$(version).exe

linux:
	env GOOS=linux GOARCH=amd64 go build -o kustomize-graph-linux-$(version)

darwin:
	env GOOS=darwin GOARCH=amd64 go build -o kustomize-graph-darwin-$(version)

all: windows linux darwin