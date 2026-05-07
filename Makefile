run:
	go run cmd/server/main.go

swag:
	swag init -g cmd/server/main.go -o ./api/docs

lint:
	golangci-lint run

dev:
	air

build:
	go build -o bin/server cmd/server/main.go

bruno:
	@bru import openapi \
		-s ./api/docs/swagger.json \
		-o /tmp/bruno-import \
		-n "Rally Backend API" \
		--collection-format bru
	@rsync -a --delete "/tmp/bruno-import/" "./api/bruno/Rally Backend API/"
	@rm -rf /tmp/bruno-import
	@cp -r ./api/bruno/_test "./api/bruno/Rally Backend API/"
	@rm -rf "./api/bruno/Rally Backend API/environments"
	@cp -r ./api/bruno/environments "./api/bruno/Rally Backend API/"
	@python3 api/bruno/inject_env.py
	@sed -i 's/mode: none/mode: bearer/' "./api/bruno/Rally Backend API/collection.bru"
	@printf '\nauth:bearer {\n  token: {{idToken}}\n}\n' >> "./api/bruno/Rally Backend API/collection.bru"

sync-api: swag bruno