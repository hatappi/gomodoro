depend:
	go get -u github.com/tcnksm/ghr
	go get -u github.com/golang/dep/cmd/dep
	go get -u github.com/jteeuwen/go-bindata/...
	dep ensure -v
bindata:
	go-bindata -pkg assets -o src/assets/assets.go assets/...
build:
	go build -o dest/gomodoro cmd/gomodoro/gomodoro.go
fmt:
	go fmt $$(go list ./... | grep -v -e 'gomodoro\/src\/assets\/' -e 'gomodoro\/vendor\/')
run:
	go run main.go
build_crosscompile_image:
	docker build -t hatappi/gomodoro-crosscompile -f docker/crosscompile/Dockerfile .
push_crosscompile_image:
	docker push hatappi/gomodoro-crosscompile
