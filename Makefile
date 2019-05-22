windows:
	env GOOS=windows GOARCH=amd64 go build -o kustomize-graph-windows-$(version).exe ./cmd/main.go

linux:
	env GOOS=linux GOARCH=amd64 go build -o kustomize-graph-linux-$(version) ./cmd/main.go

darwin:
	env GOOS=darwin GOARCH=amd64 go build -o kustomize-graph-darwin-$(version) ./cmd/main.go

all: windows linux darwin