repository = krishchow
version = v0.1

build:
	go build -o bin/main main.go

docker: build
	docker build --tag quay.io/$(repository)/ephemeral-csi-driver:$(version) .

push: docker
	docker push quay.io/$(repository)/ephemeral-csi-driver:$(version)