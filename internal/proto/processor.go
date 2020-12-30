// Copyright 2020 The Gostalkd Project Authors.
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

package proto

import (
	"errors"

	"github.com/gostalk/gostalkd/internal/network"
)

// CmdProcessor
type CmdProcessor struct {
	clients *Clients
}

// NewCmdProcessor
func NewCmdProcessor() *CmdProcessor {
	return &CmdProcessor{
		clients: clients,
	}
}

// Connect
func (p *CmdProcessor) Connect(s network.Session, r network.Reply) error {
	p.clients.PutClient(NewClient(s, r))
	return nil
}

// Process
func (p *CmdProcessor) Process(s network.Session, r network.Reply) error {
	cli, ok := p.clients.GetClientByID(s.ID())
	if !ok {
		return errors.New("client not found")
	}
	return cli.DoProcess()
}

// Destroy
func (p *CmdProcessor) Destroy(s network.Session) {
	p.clients.DelClientByID(s.ID())
}
