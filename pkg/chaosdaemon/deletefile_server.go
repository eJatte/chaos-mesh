package chaosdaemon

import (
	"context"
	"fmt"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"

	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

func (s *DaemonServer) DeleteFile(ctx context.Context, req *pb.DeleteFileRequest) (*pb.DeleteFileResponse, error) {
	log.Info("DeleteFile", "request", req)

	pid, err := s.crClient.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		log.Error(err, "Failed to get pid from container id <"+req.ContainerId+">")
		return nil, err
	}
	var attackSuccessful = true

	log.Info(fmt.Sprintf("Deleting file %s as uid %d and gid %d", req.FilePath, req.Uid, req.Gid))
	cmd := bpm.DefaultProcessBuilder("sh", "-c", fmt.Sprintf("rm -rf %s", req.FilePath)).
		SetContext(ctx).
		BuildNsEnter(pid, req.Uid, req.Gid)
	_, err = cmd.Output()
	if err != nil {
		log.Info("Failed to delete file")
		attackSuccessful = false
	}

	return &pb.DeleteFileResponse{
		AttackSuccessful: attackSuccessful,
	}, nil
}
