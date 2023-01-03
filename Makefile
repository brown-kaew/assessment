run:
	DATABASE_URL=<ChangeMe> PORT=:2565 go run server.go

sandbox:
	docker-compose -f docker-compose.test.yml down && docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit --exit-code-from it_tests