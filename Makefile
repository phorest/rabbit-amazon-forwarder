build:
	docker build -t phorest/rabbit-amazon-forwarder -f Dockerfile .

push: test build
	docker push phorest/rabbit-amazon-forwarder

test:
	docker-compose run --rm tests

dev:
	go build
