depend:
	go get -u github.com/tcnksm/ghr
	go get -u github.com/golang/dep/cmd/dep
	go get -u github.com/rakyll/statik
	go get -u golang.org/x/lint/golint
	dep ensure -v

build-assets:
	statik -src=assets -dest=libs/assets

build:
	go build -o dest/gomodoro main.go

fmt:
	go fmt $$(go list ./... | grep -v -e 'gomodoro\/libs\/assets\/' -e 'gomodoro\/vendor\/')

run:
	go run main.go

lint:
	go list ./... | xargs golint -set_exit_status

build_crosscompile_image:
	docker build -t hatappi/gomodoro-crosscompile -f docker/crosscompile/Dockerfile .

push_crosscompile_image:
	docker push hatappi/gomodoro-crosscompile
