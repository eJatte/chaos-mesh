package chaosdaemon

import (
	"context"
	"fmt"
	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	"github.com/golang/protobuf/ptypes/empty"

	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

func (s *DaemonServer) ExecHelloWorldChaos(ctx context.Context, req *pb.ExecHelloWorldRequest) (*empty.Empty, error) {
	log.Info("ExecHelloWorldChaos", "request", req)

	pid, err := s.crClient.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		log.Error(err, "Failed to get Pid from container id bla bla <"+req.ContainerId+">")
		return nil, err
	}

	cmd := bpm.DefaultProcessBuilder("sh", "-c", fmt.Sprintf("echo 'hello' `hostname`")).
		SetNS(pid, bpm.PidNS).
		SetContext(ctx).
		Build()
	out, err := cmd.Output()
	if err != nil {
		log.Error(err, "Failed to execute command")
		return nil, err
	}
	if len(out) != 0 {
		log.Info("cmd output", "output", string(out))
	}

	return &empty.Empty{}, nil
}
