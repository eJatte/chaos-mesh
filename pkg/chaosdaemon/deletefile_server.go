package chaosdaemon

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"

	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

func (s *DaemonServer) DeleteFile(ctx context.Context, req *pb.DeleteFileRequest) (*empty.Empty, error) {
	log.Info("DeleteFile", "request", req)

	log.Info("DELETE FILE path: "+req.FilePath)

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

	return &empty.Empty{}, nil
}
