## [beanstalkd-go](https://github.com/sjatsh/beanstalkd-go)

[English](readme-en.md) | 简体中文

[![Build Status](https://travis-ci.org/sjatsh/beanstalk-go.svg?branch=main)](https://travis-ci.org/sjatsh/beanstalk-go.svg?branch=main)
[![codecov](https://codecov.io/gh/sjatsh/beanstalk-go/branch/main/graph/badge.svg)](https://codecov.io/gh/sjatsh/beanstalk-go)
[![Release](https://img.shields.io/github/release/sjatsh/beanstalkd-go.svg?label=Release)](https://github.com/sjatsh/beanstalkd-go/releases)
[![License](https://img.shields.io/github/license/sjatsh/beanstalkd-go)](https://github.com/sjatsh/beanstalkd-go)

## 描述

- 简单快速的通用工作队列
- 作为学习目的为初衷，用golang完全实现了 [Beanstalk](https://github.com/beanstalkd/beanstalkd) 功能
- [协议说明](protocol.zh-CN.md)

## 里程碑

- *2020-11-14* : 所有指令全部实现完成，但仅限内存。
- *2020-11-15* : binlog持久化支持

## 快速开始

使用go get安装

```bash
go get -u -v github.com/sjatsh/beanstalkd-go
```

手动编译

```bash
make       # 编译
make run   # 运行
make test  # 运行单测
make clean # 清除编译和运行结果
```

查看支持命令

```bash
./beanstalkd-go -h
```

```bash
Usage of ./beanstalkd-go:
  -F    never fsync
  -L string
        set the log level, switch one in (panic, fatal, error, warn, waring, info, debug, trace) (default "warn")
  -V    increase verbosity
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
  -v    show version information
  -z int
        set the maximum job size in bytes (default is 65535);max allowed is 1073741824 bytes (default 65535)
```

## 第三方

- [Beanstalkd管理界面](https://github.com/xuri/aurora)
- [Beanstalkd高可用客户端](https://github.com/tal-tech/go-queue)