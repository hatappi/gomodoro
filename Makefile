bindata:
	go-bindata -pkg assets -o libs/assets/assets.go assets/...
build:
	go build -o dest/gomodoro main.go
fmt:
	go fmt $$(go list ./... | grep -v -e 'gomodoro\/libs\/assets\/' -e 'gomodoro\/vendor\/')
run:
	go run main.go
