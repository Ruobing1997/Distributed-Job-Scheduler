package dispatch

import (
	"context"
	"fmt"
	data_structure_redis "git.woa.com/robingowang/MoreFun_SuperNova/pkg/data-structure"
	"git.woa.com/robingowang/MoreFun_SuperNova/pkg/middleware"
	pb "git.woa.com/robingowang/MoreFun_SuperNova/pkg/strategy/dispatch/proto"
	generator "git.woa.com/robingowang/MoreFun_SuperNova/pkg/task-generator"
	"git.woa.com/robingowang/MoreFun_SuperNova/utils/constants"
	"google.golang.org/grpc"
	"log"
	"net"
	"time"
)

type ServerImpl struct {
	pb.UnimplementedTaskServiceServer
	pb.UnimplementedLeaseServiceServer
}

func InitManagerGRPC() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterTaskServiceServer(s, &ServerImpl{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func InitWorkerGRPC() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterLeaseServiceServer(s, &ServerImpl{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (s *ServerImpl) ExecuteTask(ctx context.Context, in *pb.TaskRequest) (*pb.TaskResponse, error) {
	payload := generator.GeneratePayload(int(in.Payload.Format), in.Payload.Script)
	task := &constants.TaskCache{
		ID:            in.Id,
		Payload:       payload,
		ExecutionTime: in.ExecutionTime.AsTime(),
		RetriesLeft:   int(in.MaxRetryCount),
	}

	err := middleware.ExecuteTaskFuncThroughMediator(task)
	if err != nil {
		return &pb.TaskResponse{Id: task.ID, Status: constants.JOBFAILED}, err
	}
	return &pb.TaskResponse{Id: task.ID, Status: constants.JOBSUCCEED}, nil
}

func (s *ServerImpl) RenewLease(ctx context.Context, in *pb.RenewLeaseRequest) (*pb.RenewLeaseResponse, error) {
	// worker should ask for a new lease before the current lease expires
	err := data_structure_redis.SetLeaseWithID(in.Id, 2*time.Second)
	if err != nil {
		return &pb.RenewLeaseResponse{Success: false},
			fmt.Errorf("lease for update %s has expired", in.Id)
	}
	return &pb.RenewLeaseResponse{Success: true}, nil
}

func (s *ServerImpl) StopTask(ctx context.Context, in *pb.TaskStopRequest) (*pb.TaskResponse, error) {
	// Similar with ExecuteTask, but it is to stop a update
	return nil, nil // TODO
}

func StopRunningTask(runningTask *constants.RunTimeTask) (string, int, error) {
	// TODO
	return "", 0, nil
}
