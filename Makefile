APP_NAME=karen
REGISTRY?=343660461351.dkr.ecr.us-east-2.amazonaws.com
TAG?=latest
KAREN_DB_LOCATION?=localhost:3306
KAREN_DB_USERNAME?=karen
KAREN_DB_PASSWORD?=karen
KAREN_DB_DATABASE?=karen
HELP_FUNC = \
    %help; \
    while(<>) { \
        if(/^([a-z0-9_-]+):.*\#\#(?:@(\w+))?\s(.*)$$/) { \
            push(@{$$help{$$2 // 'targets'}}, [$$1, $$3]); \
        } \
    }; \
    print "usage: make [target]\n\n"; \
    for ( sort keys %help ) { \
        print "$$_:\n"; \
        printf("  %-20s %s\n", $$_->[0], $$_->[1]) for @{$$help{$$_}}; \
        print "\n"; \
    }

.PHONY: help
help: 				## show options and their descriptions
	@perl -e '$(HELP_FUNC)' $(MAKEFILE_LIST)

all:  				## clean the working environment, build and test the packages, and then build the docker image
all: clean test docker

tmp: 				## create tmp/
	if [ -d "./tmp" ]; then rm -rf ./tmp; fi
	mkdir tmp

build: tmp 			## build the app binaries
	go build -o ./tmp ./...

test: build 		## build and test the module packages
	go test ./...

run: build 			## build and run the app binaries
	export KAREN_DB_LOCATION=$(KAREN_DB_LOCATION) && \
		export KAREN_DB_USERNAME=$(KAREN_DB_USERNAME) && \
		export KAREN_DB_PASSWORD=$(KAREN_DB_PASSWORD) && \
		export KAREN_DB_DATABASE=$(KAREN_DB_DATABASE) && \
		./tmp/app

docker: tmp 		## build the docker image
	wget -O tmp/rds-combined-ca-bundle.pem https://s3.amazonaws.com/rds-downloads/rds-combined-ca-bundle.pem
	docker build -t $(REGISTRY)/$(APP_NAME):$(TAG) .

docker-run: docker 	## start the built docker image in a container
	docker run -d -p 80:80 \
		-e KAREN_DB_LOCATION=$(KAREN_DB_LOCATION) \
		-e KAREN_DB_USERNAME=$(KAREN_DB_USERNAME) \
		-e KAREN_DB_PASSWORD=$(KAREN_DB_PASSWORD) \
		-e KAREN_DB_DATABASE=$(KAREN_DB_DATABASE) \
		--name $(APP_NAME) $(REGISTRY)/$(APP_NAME):$(TAG)

docker-push: tmp docker
	docker push $(REGISTRY)/$(APP_NAME):$(TAG)

.PHONY: clean
clean: 				## remove tmp/, stop and remove app container, old docker images
	rm -rf tmp
ifneq ("$(shell docker container list -a | grep $(APP_NAME))", "")
	docker rm -f $(APP_NAME)
endif
	docker system prune
ifneq ("$(shell docker images | grep $(APP_NAME) | awk '{ print $$3; }')", "") 
	docker images | grep $(APP_NAME) | awk '{ print $$3; }' | xargs -I {} docker rmi -f {}
endif
