## beanstalk-go

English | [简体中文](readme.md)

[![Build Status](https://travis-ci.org/sjatsh/beanstalk-go.svg?branch=main)](https://travis-ci.org/sjatsh/beanstalk-go.svg?branch=main)
[![Codecov](https://img.shields.io/codecov/c/github/sjatsh/beanstalk-go/main)](https://github.com/sjatsh/beanstalk-go)
[![Release](https://img.shields.io/github/release/sjatsh/beanstalk-go.svg?label=Release)](https://github.com/sjatsh/beanstalk-go/releases)
[![License](https://img.shields.io/github/license/sjatsh/beanstalk-go)](https://github.com/sjatsh/beanstalk-go)

## Description

 - Simple and fast general purpose work queue.
 - Fully [Beanstalk](https://github.com/beanstalkd/beanstalkd) compatible task queue implemented by golang for learning
purpose
 - [ProtocolDescription](protocol.zh-CN.md)

## Milepost

 - *2020-11-14* : all dispatch cmd complete but memory only, binlog period not supported. 

## Quick Start

using go get install
```bash
GO111MODULE=on GOPROXY=https://goproxy.cn/,direct go get -u -v github.com/sjatsh/beanstalk-go
```

using make you self
```bash
make
./beanstalkd
```
view support commands

```bash
./beanstalk -h
```

```bash
Usage of ./beanstalkd:
-F	never fsync
-V	increase verbosity
-b string
  	write-ahead log directory
-f int
  	fsync at most once every MS milliseconds (default is 50ms);use -f0 for "always fsync" (default 50)
-l string
  	listen on address (default is 0.0.0.0) (default "0.0.0.0")
-p int
  	listen on port (default is 11400) (default 11400)
-s int
  	set the size of each write-ahead log file (default is 10485760);will be rounded up to a multiple of 4096 bytes (default 10485760)
-u string
  	become user and group
-v	show version information
-z int
  	set the maximum job size in bytes (default is 65535);max allowed is 1073741824 bytes (default 65535)
```

## Third Party

 - [Beanstalkd queue server console](https://github.com/xuri/aurora)
 - [High available beanstalkd go client](https://github.com/tal-tech/go-queue)