// Copyright 2020 SunJun <i@sjis.me>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package model

import (
	"github.com/gostalk/gostalkd/structure"
)

type ServerOptions struct {
	Port int // 服务端口
	Addr string
	User string
}

type ServerOption func(*ServerOptions)

type Server struct {
	Options ServerOptions
	Wal     Wal

	Sock   *Socket
	Connes *structure.Heap
}

// WithPort 设置端口
func WithPort(port int) ServerOption {
	return func(o *ServerOptions) {
		o.Port = port
	}
}

// WithAddr 设置服务端监听ip
func WithAddr(addr string) ServerOption {
	return func(o *ServerOptions) {
		o.Addr = addr
	}
}

// WithUser 运行用户
func WithUser(user string) ServerOption {
	return func(o *ServerOptions) {
		o.User = user
	}
}
