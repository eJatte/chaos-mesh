package chaosdaemon

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"

	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

func (s *DaemonServer) ExecHelloWorldChaos(ctx context.Context, req *pb.ExecHelloWorldRequest) (*empty.Empty, error) {
	log.Info("ExecHelloWorldChaos", "request", req)

	pid, err := s.crClient.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		log.Error(err, "Failed to get Pid from container id bla bla <"+req.ContainerId+">")
		return nil, err
	}

	log.Info("Executing ls")
	cmd := bpm.DefaultProcessBuilder("sh", "-c", fmt.Sprintf("ls")).
		SetNS(pid, bpm.PidNS).
		SetContext(ctx).
		Build()
	out, err := cmd.Output()
	if err != nil {
		log.Error(err, "Failed to execute ls")
		return nil, err
	}
	if len(out) != 0 {
		log.Info("cmd output", "output", string(out))
	}

	log.Info("Executing echo hello world to file")
	cmd = bpm.DefaultProcessBuilder("sh", "-c", fmt.Sprintf("echo \"hello world!\" >> hello.txt")).
		SetNS(pid, bpm.PidNS).
		SetContext(ctx).
		Build()
	out, err = cmd.Output()
	if err != nil {
		log.Error(err, "Failed to execute echo")
		return nil, err
	}
	if len(out) != 0 {
		log.Info("cmd output", "output", string(out))
	}

	log.Info("Executing ls")
	cmd = bpm.DefaultProcessBuilder("sh", "-c", fmt.Sprintf("ls")).
		SetNS(pid, bpm.PidNS).
		SetContext(ctx).
		Build()
	out, err = cmd.Output()
	if err != nil {
		log.Error(err, "Failed to execute ls")
		return nil, err
	}
	if len(out) != 0 {
		log.Info("cmd output", "output", string(out))
	}

	log.Info("Executing cat")
	cmd = bpm.DefaultProcessBuilder("sh", "-c", fmt.Sprintf("cat hello.txt")).
		SetNS(pid, bpm.PidNS).
		SetContext(ctx).
		Build()

	out, err = cmd.Output()
	if err != nil {
		log.Error(err, "Failed to execute cat")
		return nil, err
	}
	if len(out) != 0 {
		log.Info("cmd output", "output", string(out))
	}

	log.Info("Executing rm")
	cmd = bpm.DefaultProcessBuilder("sh", "-c", fmt.Sprintf("rm hello.txt")).
		SetNS(pid, bpm.PidNS).
		SetContext(ctx).
		Build()
	out, err = cmd.Output()
	if err != nil {
		log.Error(err, "Failed to execute rm")
		return nil, err
	}
	if len(out) != 0 {
		log.Info("cmd output", "output", string(out))
	}

	log.Info("Executing ls")
	cmd = bpm.DefaultProcessBuilder("sh", "-c", fmt.Sprintf("ls")).
		SetNS(pid, bpm.PidNS).
		SetContext(ctx).
		Build()
	out, err = cmd.Output()
	if err != nil {
		log.Error(err, "Failed to execute ls")
		return nil, err
	}
	if len(out) != 0 {
		log.Info("cmd output", "output", string(out))
	}

	return &empty.Empty{}, nil
}
