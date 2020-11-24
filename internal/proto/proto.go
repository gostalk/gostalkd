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
	"time"

	"github.com/gostalk/gostalkd/internal/utils/str"
)

func ioProcess(c *Client) error {
	switch c.state {
	case StateWantCommand:
		return dispatchCmd(c)
	case StateWantData:
		l := int64(len(c.session.Data()))
		if c.inJob.R.BodySize != l {
			c.reply.WriteString(MsgBadFormat)
			c.reply.Flush()
			return nil
		}
		c.inJob.body = c.session.Data()

		enqueueJob(c)

		c.inJob = nil
		c.state = StateWantCommand
	}
	return nil
}

// dispatchCmd
func dispatchCmd(c *Client) error {
	cmd := string(c.session.Data())
	tokens := str.Parse(cmd)
	cmdName, err := tokens.NextString()
	if err != nil {
		c.reply.WriteString(MsgBadFormat)
		c.reply.Flush()
		return err
	}

	option, ok := Cmd2Option[cmdName]
	if !ok {
		c.reply.WriteString(MsgNotFound)
		c.reply.Flush()
		return nil
	}

	switch option {
	case OpPut:
		dispatchPut(c, tokens)
	}

	return nil
}

// dispatchPut
func dispatchPut(c *Client, tokens *str.Tokens) {
	if tokens.Len() < 5 {
		c.reply.WriteString(MsgBadFormat)
		c.reply.Flush()
		return
	}
	pri, err := tokens.NextUint32()
	if err != nil {
		c.reply.WriteString(MsgBadFormat)
		c.reply.Flush()
		return
	}
	delay, err := tokens.NextInt64()
	if err != nil {
		c.reply.WriteString(MsgBadFormat)
		c.reply.Flush()
		return
	}
	ttr, err := tokens.NextInt64()
	if err != nil {
		c.reply.WriteString(MsgBadFormat)
		c.reply.Flush()
		return
	}
	bodySize, err := tokens.NextInt64()
	if err != nil {
		c.reply.WriteString(MsgBadFormat)
		c.reply.Flush()
		return
	}

	if ttr < 1000000000 {
		ttr = 1000000000
	}

	j := NewJob().WithPri(pri).
		WithDelay(delay).
		WithTTR(ttr).
		WithBodySize(bodySize).
		WithTube(c.use)

	c.inJob = j
	c.state = StateWantData
}

// enqueueJob
func enqueueJob(c *Client) {
	j := c.inJob
	if j.R.Delay > 0 {
		j.R.State = JobDelayed
		j.R.DeadlineAt = time.Now().UnixNano() + c.inJob.R.Delay
		j.tube.PushDelayJob(c.inJob.Cover2DelayJob())
	} else {
		j.R.State = JobReady
		j.GetTube().PushReadyJob(j.Cover2ReadyJob())
	}
}
