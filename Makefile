NAME=beanstalkd-go
VERSION=$(shell ./vers.sh)
GOHOSTOS=$(shell go env GOHOSTOS)

INSTALL_DIR=/usr/local/bin/
LOG_DIR=/usr/local/var/${NAME}/log

all:
	GOOS=${GOHOSTOS} go build  -ldflags "-X 'main.version=${VERSION}'" -o ${NAME} .
darwin:
	GOOS=darwin go build  -ldflags "-X 'main.version=${VERSION}'" -o ${NAME} .
linux:
	GOOS=linux go build  -ldflags "-X 'main.version=${VERSION}'" -o ${NAME} .
install:
	cp -f ${NAME} ${INSTALL_DIR}
	mkdir -p ${LOG_DIR}
uninstall:
	rm -rf ${INSTALL_DIR}${NAME}
	rm -rf ${LOG_DIR}
run:
	${INSTALL_DIR}${NAME} -b ${LOG_DIR}
clean:
	rm -rf ${NAME}