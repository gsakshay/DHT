# Define variables
BOOTSTRAP_IMAGE_NAME = prj5-bootstrap
PEER_IMAGE_NAME = prj5-peer
CLIENT_IMAGE_NAME =  prj5-client

COMPOSE_TEST1 = docker-compose-testcase-1.yml
COMPOSE_TEST2 = docker-compose-testcase-2.yml
COMPOSE_TEST3 = docker-compose-testcase-3.yml
COMPOSE_TEST4 = docker-compose-testcase-4.yml
COMPOSE_TEST5 = docker-compose-testcase-5.yml

# Default target: build the Docker image
.PHONY: build
build:
	docker build -t $(BOOTSTRAP_IMAGE_NAME) -t $(PEER_IMAGE_NAME) -t $(CLIENT_IMAGE_NAME) .

# Run the first test case
.PHONY: test1
test1: build
	docker compose -f $(COMPOSE_TEST1) up --build

# Run the second test case
.PHONY: test2
test2: build
	docker compose -f $(COMPOSE_TEST2) up --build

# Run the third test case
.PHONY: test3
test3: build
	docker compose -f $(COMPOSE_TEST3) up --build

# Run the fourth test case
.PHONY: test4
test4: build
	docker compose -f $(COMPOSE_TEST4) up --build

# Run the fifth test case
.PHONY: test5
test5: build
	docker compose -f $(COMPOSE_TEST5) up --build

# Stop and remove containers for test1
.PHONY: down-test1
down-test1:
	docker compose -f $(COMPOSE_TEST1) down

# Stop and remove containers for test2
.PHONY: down-test2
down-test2:
	docker compose -f $(COMPOSE_TEST2) down

# Stop and remove containers for test3
.PHONY: down-test3
down-test3:
	docker compose -f $(COMPOSE_TEST3) down

# Stop and remove containers for test4
.PHONY: down-test4
down-test4:
	docker compose -f $(COMPOSE_TEST4) down

# Stop and remove containers for test5
.PHONY: down-test5
down-test5:
	docker compose -f $(COMPOSE_TEST5) down

# Clean up all containers, images, and networks
.PHONY: clean
clean: down-test1 down-test2 down-test3 down-test4 down-test5
	docker image rm $(IMAGE_NAME) || true
	docker volume prune -f
	docker network prune -f
	docker container prune -f

# Run all the test cases sequentially
.PHONY: test
test: test1 down-test1 test2 down-test2 test3 down-test3 test4 down-test4 test5 down-test5

