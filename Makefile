
VERSION=$(shell ./vers.sh)

all:
	go build  -ldflags "-X 'main.version=${VERSION}'" -o beanstalk-go .