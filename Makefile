NAME="beanstalk-go"
VERSION=$(shell ./vers.sh)

all:
	go build  -ldflags "-X 'main.version=${VERSION}'" -o ${NAME} .
darwin:
	GOOS=darwin go build  -ldflags "-X 'main.version=${VERSION}'" -o ${NAME} .
linux:
	GOOS=linux go build  -ldflags "-X 'main.version=${VERSION}'" -o ${NAME} .
clean:
	rm -rf ${NAME}