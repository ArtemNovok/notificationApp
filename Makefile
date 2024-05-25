build:
	@echo Stopping running containers...
	docker-compose down -v
	@echo Building new images if required and startin containters...
	docker-compose up --build -d
	@echo Images builded and started!

stop:
	@echo Stopping running containers...
	docker-compose down -v
	@echo Containers are stopped!