bindata:
	go-bindata -pkg assets -o libs/assets/assets.go assets/...
build:
	go build -o dest/gomodoro main.go
run:
	go run main.go

