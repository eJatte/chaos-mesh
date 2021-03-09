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

	pid, err := s.crClient.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		log.Error(err, "Failed to get Pid from container id bla bla <"+req.ContainerId+">")
		return nil, err
	}
	fileName := "dummyfile"
	filePath := fmt.Sprintf("%s%s", req.DirectoryPath, fileName)

	log.Info(fmt.Sprintf("Creating file %s", filePath))
	cmd := bpm.DefaultProcessBuilder("sh", "-c", fmt.Sprintf("echo 'hello friend' >> %s", filePath)).
		SetContext(ctx).
		BuildNsEnter(pid, 0)
	out, err := cmd.Output()
	if err != nil {
		log.Error(err, "Failed to create file.")
		return nil, err
	}

	log.Info("Executing ls")
	cmd = bpm.DefaultProcessBuilder("sh", "-c", fmt.Sprintf("cd %s && ls", req.DirectoryPath)).
		SetContext(ctx).
		BuildNsEnter(pid, 0)
	out, err = cmd.Output()
	if err != nil {
		log.Error(err, "Failed to execute ls")
	}
	if len(out) != 0 {
		log.Info("cmd output", "output", string(out))
	}

	log.Info(fmt.Sprintf("Deleting file %s as user %d", filePath, req.Uid))
	cmd = bpm.DefaultProcessBuilder("sh", "-c", fmt.Sprintf("rm -rf %s", filePath)).
		SetContext(ctx).
		BuildNsEnter(pid, req.Uid)
	out, err = cmd.Output()
	if err != nil {
		log.Info("Failed to delete file")
	}

	log.Info(fmt.Sprintf("Deleting file %s", filePath))
	cmd = bpm.DefaultProcessBuilder("sh", "-c", fmt.Sprintf("rm -rf %s", filePath)).
		SetContext(ctx).
		BuildNsEnter(pid, 0)
	out, err = cmd.Output()
	if err != nil {
		log.Error(err, "Failed to delete file as root")
	}

	log.Info("Executing ls")
	cmd = bpm.DefaultProcessBuilder("sh", "-c", fmt.Sprintf("cd %s && ls", req.DirectoryPath)).
		SetContext(ctx).
		BuildNsEnter(pid, 0)
	out, err = cmd.Output()
	if err != nil {
		log.Error(err, "Failed to execute ls")
	}
	if len(out) != 0 {
		log.Info("cmd output", "output", string(out))
	}

	return &empty.Empty{}, nil
}
