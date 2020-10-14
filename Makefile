
build:
	go build -o bin/main main.go

docker: build
	docker build --tag quay.io/krishchow/ephemeral-csi-driver:v0.06 .

push: docker
	docker push quay.io/krishchow/ephemeral-csi-driver:v0.06