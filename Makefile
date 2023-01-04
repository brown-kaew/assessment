run:
	DATABASE_URL=ChangeMe PORT=:2565 go run server.go

unit:
	go test -v ./...

sandbox:
	docker-compose -f docker-compose.test.yml down && docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit --exit-code-from it_tests

docker-build:
	docker build -t assessment:app .

docker-run:
	docker run -e DATABASE_URL -e PORT -p 2565:2565 assessment:app
