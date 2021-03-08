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
		SetContext(ctx).
		BuildNsEnter(pid,0)
	out, err := cmd.Output()
	if err != nil {
		log.Error(err, "Failed to execute ls")
		return nil, err
	}
	if len(out) != 0 {
		log.Info("cmd output", "output", string(out))
	}

	log.Info("Executing whoami")
	cmd = bpm.DefaultProcessBuilder("sh", "-c", fmt.Sprintf("whoami")).
		SetContext(ctx).
		BuildNsEnter(pid,0)
	out, err = cmd.Output()
	if err != nil {
		log.Error(err, "Failed to execute whoami")
		return nil, err
	}
	if len(out) != 0 {
		log.Info("cmd output", "output", string(out))
	}

	log.Info("Executing echo uid")
	cmd = bpm.DefaultProcessBuilder("sh", "-c", fmt.Sprintf("echo $UID")).
		SetContext(ctx).
		BuildNsEnter(pid,1000)
	out, err = cmd.Output()
	if err != nil {
		log.Error(err, "Failed to execute echo")
		return nil, err
	}
	if len(out) != 0 {
		log.Info("cmd output", "output", string(out))
	}

	log.Info("Executing whoami as user 1000")
	cmd = bpm.DefaultProcessBuilder("sh", "-c", fmt.Sprintf("whoami")).
		SetContext(ctx).
		BuildNsEnter(pid,1000)
	out, err = cmd.Output()
	if err != nil {
		log.Error(err, "Failed to execute whoami as user 1000")
	}
	if len(out) != 0 {
		log.Info("cmd output", "output", string(out))
	}

	log.Info("Executing echo as user 1000")
	cmd = bpm.DefaultProcessBuilder("sh", "-c", fmt.Sprintf("echo 'hello' `hostname`")).
		SetContext(ctx).
		BuildNsEnter(pid,1000)
	out, err = cmd.Output()
	if err != nil {
		log.Error(err, "Failed to execute echo as user 1000")
	}
	if len(out) != 0 {
		log.Info("cmd output", "output", string(out))
	}

	log.Info("Executing cd write to file as user 1000")
	cmd = bpm.DefaultProcessBuilder("sh", "-c", fmt.Sprintf("echo 'hello friend' >> hello.txt")).
		SetContext(ctx).
		BuildNsEnter(pid, 1000)
	out, err = cmd.Output()
	if err != nil {
		log.Error(err, "Failed to cd write to file as user 1000")
	}
	if len(out) != 0 {
		log.Info("cmd output", "output", string(out))
	}

	log.Info("Executing echo")
	cmd = bpm.DefaultProcessBuilder("sh", "-c", fmt.Sprintf("echo 'hello' `hostname`")).
		SetContext(ctx).
		BuildNsEnter(pid,0)
	out, err = cmd.Output()
	if err != nil {
		log.Error(err, "Failed to execute echo")
		return nil, err
	}
	if len(out) != 0 {
		log.Info("cmd output", "output", string(out))
	}

	log.Info("Executing cd write to file")
	cmd = bpm.DefaultProcessBuilder("sh", "-c", fmt.Sprintf("cd super/data/ && echo 'hello friend' >> hello.txt")).
		SetContext(ctx).
		BuildNsEnter(pid, 0)
	out, err = cmd.Output()
	if err != nil {
		log.Error(err, "Failed to cd write to file")
		return nil, err
	}
	if len(out) != 0 {
		log.Info("cmd output", "output", string(out))
	}

	log.Info("Executing removing file")
	cmd = bpm.DefaultProcessBuilder("sh", "-c", fmt.Sprintf("rm -rf "+req.FilePath)).
		SetContext(ctx).
		BuildNsEnter(pid, 0)
	out, err = cmd.Output()
	if err != nil {
		log.Error(err, "Failed to delete file")
		return nil, err
	}
	if len(out) != 0 {
		log.Info("cmd output", "output", string(out))
	}

	return &empty.Empty{}, nil
}
