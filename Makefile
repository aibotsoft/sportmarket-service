#source .env
.EXPORT_ALL_VARIABLES:
SERVICE_NAME=sportmarket-service
DOCKER_USERNAME=aibotsoft
CGO_ENABLED=0
GOARCH=amd64

linux_build:
	GOOS=linux go build -o dist/service main.go

build:
	go build -o dist/service main.go

run:
	go run main.go

test:
	SERVICE_ENV=test
	go test -v -cover ./...

run_build:
	dist/service

#Команды для докера
docker_build:
	docker image build -f Dockerfile -t $$DOCKER_USERNAME/$$SERVICE_NAME .

docker_run_rm:
	docker run --rm -t $$DOCKER_USERNAME/$$SERVICE_NAME

docker_login:
	docker login -u $$DOCKER_USERNAME -p $$DOCKER_PASSWORD

docker_push:
	docker push $$DOCKER_USERNAME/$$SERVICE_NAME

docker_deploy: linux_build docker_build docker_login docker_push

#Команды для k8s
kube_deploy:
	kubectl apply -f k8s/

kube_rol:
	kubectl -n micro rollout restart deployment $$SERVICE_NAME

mig_up:
	source .env
	$$DSL=sqlserver://$$MSSQL_USER:$$MSSQL_PASSWORD@$$mssql_host:$$mssql_port?database=$$MSSQL_DATABASE
	migrate -verbose -source file://migrations -database $$DSL goto 2

mig_create:
	migrate create -ext sql -dir migrations uspSaveEvent
