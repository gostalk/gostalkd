//go:build windows
// +build windows

// Copyright 2020 gostalkd
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

package net

import (
	"time"

	"github.com/gostalk/gostalkd/model"
)

func init() {
}

func sockWant(s *model.Socket, rw byte) error {
	return nil
}

func sockNext(s **model.Socket, timeout time.Duration) (byte, error) {
	return 0, nil
}
