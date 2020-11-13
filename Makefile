
VERSION=$(shell ./version.sh)

all:
	go build  -ldflags "-X 'main.version=${VERSION}'" -o beanstalkd .