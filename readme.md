## [gostalkd](https://github.com/gostalk/gostalkd)

English | [简体中文](readme.md)

[![Build Status](https://travis-ci.org/gostalk/gostalkd.svg?branch=main)](https://travis-ci.org/gostalk/gostalkd.svg?branch=main)
[![codecov](https://codecov.io/gh/gostalk/gostalkd/branch/main/graph/badge.svg)](https://codecov.io/gh/gostalk/gostalkd)
[![Release](https://img.shields.io/github/release/gostalk/gostalkd.svg?label=Release)](https://github.com/gostalk/gostalkd/releases)
[![License](https://img.shields.io/github/license/gostalk/gostalkd)](https://github.com/gostalk/gostalkd)

## Description

- Simple and fast general purpose work queue.
- Fully compatible with beanstalkd protocol
- Fully [Beanstalk](https://github.com/beanstalkd/beanstalkd) compatible task queue implemented by golang
  purpose
- [ProtocolDescription](doc/protocol.zh-CN.md)
- **P0 fixes**: CRLF detection, reserve-job NOT_FOUND for reserved jobs
- **P1 features**: connection limit (-c), graceful shutdown (-t), drain command

## Milepost

- *2020-11-14* : all dispatch cmd complete but memory only.
- *2020-11-15* : binlog persistence support
- *2026-03-18* : P0 bug fixes + connection limit, graceful shutdown, drain command

## Quick Start

using go get install

```bash
GO111MODULE=on GOPROXY=https://goproxy.cn/,direct go get -u -v github.com/gostalk/gostalkd
```

using make you self

```bash
make
make run    # run program
make test   # run go test
make clean  # del program and log dir
```

view support commands

```bash
./gostalkd -h
```

```bash
Usage of ./gostalkd:
  -c int
        set the maximum number of concurrent connections (default is 0, meaning no limit)
  -F    never fsync
  -f int
        fsync at most once every MS milliseconds (default is 50ms);use -f0 for "always fsync" (default 50)
  -l string
        listen on address (default is 0.0.0.0) (default "0.0.0.0")
  -p int
        listen on port (default is 11400) (default 11400)
  -s int
        set the size of each write-ahead log file (default is 10485760);will be rounded up to a multiple of 4096 bytes (default 10485760)
  -t int
        set the graceful shutdown timeout in seconds (default is 30)
  -u string
        become user and group
  -v    show version information
  -z int
        set the maximum job size in bytes (default is 65535);max allowed is 1073741824 bytes (default 65535)
```

## Third Party

- [Beanstalkd queue server console](https://github.com/xuri/aurora)
- [High available beanstalkd go client](https://github.com/tal-tech/go-queue) 