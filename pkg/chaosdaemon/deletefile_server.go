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

	log.Info("Creating file")
	cmd := bpm.DefaultProcessBuilder("sh", "-c", fmt.Sprintf("echo 'hello friend' >> super/data/dummyfile")).
		SetContext(ctx).
		BuildNsEnter(pid, 0)
	out, err := cmd.Output()
	if err != nil {
		log.Error(err, "Failed to create file.")
		return nil, err
	}

	log.Info("Executing ls")
	cmd = bpm.DefaultProcessBuilder("sh", "-c", fmt.Sprintf("cd super/data/ && ls")).
		SetContext(ctx).
		BuildNsEnter(pid,0)
	out, err = cmd.Output()
	if err != nil {
		log.Error(err, "Failed to execute ls")
	}
	if len(out) != 0 {
		log.Info("cmd output", "output", string(out))
	}

	log.Info("Deleting file as user 1000")
	cmd = bpm.DefaultProcessBuilder("sh", "-c", fmt.Sprintf("rm -rf super/data/dummyfile")).
		SetContext(ctx).
		BuildNsEnter(pid,1000)
	out, err = cmd.Output()
	if err != nil {
		log.Error(err, "Failed to delete file as user 1000")
	}

	log.Info("Deleting file as root")
	cmd = bpm.DefaultProcessBuilder("sh", "-c", fmt.Sprintf("rm -rf super/data/dummyfile")).
		SetContext(ctx).
		BuildNsEnter(pid,0)
	out, err = cmd.Output()
	if err != nil {
		log.Error(err, "Failed to delete file as root")
	}

	log.Info("Executing ls")
	cmd = bpm.DefaultProcessBuilder("sh", "-c", fmt.Sprintf("cd super/data/ && ls")).
		SetContext(ctx).
		BuildNsEnter(pid,0)
	out, err = cmd.Output()
	if err != nil {
		log.Error(err, "Failed to execute ls")
	}
	if len(out) != 0 {
		log.Info("cmd output", "output", string(out))
	}

	return &empty.Empty{}, nil
}
