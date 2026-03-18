## [gostalkd](https://github.com/gostalk/gostalkd)

[English](readme-en.md) | 简体中文

[![Build Status](https://travis-ci.org/gostalk/gostalkd.svg?branch=main)](https://travis-ci.org/gostalk/gostalkd.svg?branch=main)
[![codecov](https://codecov.io/gh/gostalk/gostalkd/branch/main/graph/badge.svg)](https://codecov.io/gh/gostalk/gostalkd)
[![Release](https://img.shields.io/github/release/gostalk/gostalkd.svg?label=Release)](https://github.com/gostalk/gostalkd/releases)
[![License](https://img.shields.io/github/license/gostalk/gostalkd)](https://github.com/gostalk/gostalkd)

## 描述

- 简单快速的通用工作队列
- 完全兼容beanstalkd协议
- 用golang完全实现了 [Beanstalk](https://github.com/beanstalkd/beanstalkd) 功能
- [协议说明](doc/protocol.zh-CN.md)
- **P0 修复**: CRLF 检测、reserve-job 对 reserved 状态返回 NOT_FOUND
- **P1 功能**: 连接数限制 (-c)、优雅关闭 (-t)、drain 命令

## 里程碑

- *2020-11-14* : 所有指令全部实现完成，但仅限内存。
- *2020-11-15* : binlog持久化支持
- *2026-03-18* : P0 修复 + 连接数限制、优雅关闭、drain 命令

## 快速开始

使用go get安装

```bash
GO111MODULE=on GOPROXY=https://goproxy.cn/,direct go get -u -v github.com/gostalk/gostalkd
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
./gostalkd -h
```

```bash
Usage of ./gostalkd:
  -c int
        最大并发连接数（默认 0，表示无限制）
  -F    禁用 fsync
  -L string
        设置日志级别，可选值: panic, fatal, error, warn, waring, info, debug, trace（默认 "warn"）
  -V    增加详细输出
  -b string
        write-ahead log 目录
  -f int
        每次写入后最多延迟 fsync 多少毫秒（默认 50ms）；使用 -f0 表示"始终 fsync"（默认 50）
  -l string
        监听地址（默认 0.0.0.0）（默认 "0.0.0.0"）
  -p int
        监听端口（默认 11400）（默认 11400）
  -s int
        每个 write-ahead log 文件的大小（默认 10485760），会被取整为 4096 的倍数（默认 10485760）
  -t int
        优雅关闭超时时间，单位秒（默认 30）
  -u string
        切换运行用户和用户组
  -v    显示版本信息
  -z int
        最大 job 大小，单位字节（默认 65535），最大允许 1073741824（默认 65535）
```

## 第三方

- [Beanstalkd管理界面](https://github.com/xuri/aurora)
- [Beanstalkd高可用客户端](https://github.com/tal-tech/go-queue) 