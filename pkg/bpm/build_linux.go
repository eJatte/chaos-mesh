// Copyright 2020 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package bpm

import (
	"context"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

// Build builds the process
func (b *ProcessBuilder) Build() *ManagedProcess {
	args := b.args
	cmd := b.cmd

	if len(b.nsOptions) > 0 {
		args = append([]string{"--", cmd}, args...)
		for _, option := range b.nsOptions {
			args = append([]string{"-" + nsArgMap[option.Typ], option.Path}, args...)
		}

		if b.localMnt {
			args = append([]string{"-l"}, args...)
		}
		cmd = nsexecPath
	}

	if b.pause {
		args = append([]string{cmd}, args...)
		cmd = pausePath
	}

	if c := mock.On("MockProcessBuild"); c != nil {
		f := c.(func(context.Context, string, ...string) *exec.Cmd)
		return &ManagedProcess{
			Cmd:        f(b.ctx, cmd, args...),
			Identifier: b.identifier,
		}
	}

	log.Info("build command", "command", cmd+" "+strings.Join(args, " "))

	command := exec.CommandContext(b.ctx, cmd, args...)
	command.SysProcAttr = &syscall.SysProcAttr{}
	command.SysProcAttr.Pdeathsig = syscall.SIGTERM

	return &ManagedProcess{
		Cmd:        command,
		Identifier: b.identifier,
	}
}

func (b *ProcessBuilder) BuildNsEnter(pid uint32, uid int64, gid int64) *ManagedProcess {
	args := b.args
	cmd := b.cmd

	args = append([]string{"-G", strconv.Itoa(int(gid)), "-S", strconv.Itoa(int(uid)), "--target", strconv.Itoa(int(pid)), "--mount", "--uts", "--ipc", "--net", "--pid", "--", cmd}, args...)

	cmd = "nsenter"

	log.Info("build command", "command", cmd+" "+strings.Join(args, " "))

	command := exec.CommandContext(b.ctx, cmd, args...)
	command.SysProcAttr = &syscall.SysProcAttr{}

	return &ManagedProcess{
		Cmd:        command,
		Identifier: b.identifier,
	}
}
