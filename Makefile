IMAGE=excelsiorio/meuip.io:latest

build:
	docker build -t $(IMAGE) .

push: build
	docker push $(IMAGE)