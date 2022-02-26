
build:
	docker build -t docker.io/mhy101/pod-webhook-example:latest -f Dockerfile .

push: build
	docker push docker.io/mhy101/pod-webhook-example:latest

.PHONY: build