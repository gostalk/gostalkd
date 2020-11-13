## beanstalk-go

[English](readme-en.md) | 简体中文

[![Build Status](https://travis-ci.org/sjatsh/beanstalk-go.svg?branch=main)](https://travis-ci.org/sjatsh/beanstalk-go.svg?branch=main)

## 描述

 - 简单快速的通用工作队列
 - 作为学习目的为初衷，用golang完全实现了 [Beanstalk](https://github.com/beanstalkd/beanstalkd) 功能
 - [协议说明](protocol.zh-CN.md)

## 里程碑

- *2020-11-14* : 所有指令全部实现完成，但仅限内存，短时间暂不支持binlog持久化。

## 快速开始

```bash
make
./beanstalkd
```

查看支持命令

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

## 第三方

- [Beanstalkd管理界面](https://github.com/xuri/aurora)
- [Beanstalkd高可用客户端](https://github.com/tal-tech/go-queue)