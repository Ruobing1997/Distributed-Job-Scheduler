package dispatch

import (
	"context"
	"fmt"
	data_structure_redis "git.woa.com/robingowang/MoreFun_SuperNova/pkg/data-structure"
	"git.woa.com/robingowang/MoreFun_SuperNova/pkg/middleware"
	pb "git.woa.com/robingowang/MoreFun_SuperNova/pkg/strategy/dispatch/proto"
	task_executor "git.woa.com/robingowang/MoreFun_SuperNova/pkg/task-executor"
	generator "git.woa.com/robingowang/MoreFun_SuperNova/pkg/task-generator"
	"git.woa.com/robingowang/MoreFun_SuperNova/utils/constants"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"net"
	"time"
)

type ServerImpl struct {
	pb.UnimplementedTaskServiceServer
	pb.UnimplementedLeaseServiceServer
}

func Init() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterTaskServiceServer(s, &ServerImpl{})
	pb.RegisterLeaseServiceServer(s, &ServerImpl{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	middleware.SetRenewLeaseFunction(RenewLease)
}

func (s *ServerImpl) ExecuteTask(ctx context.Context, in *pb.TaskRequest) (*pb.TaskResponse, error) {
	payload := generator.GeneratePayload(int(in.Payload.Format), in.Payload.Script)
	task := &constants.TaskCache{
		ID:            in.Id,
		Payload:       payload,
		ExecutionTime: in.ExecutionTime.AsTime(),
		RetriesLeft:   int(in.MaxRetryCount),
	}

	err := task_executor.ExecuteTask(task)
	if err != nil {
		return &pb.TaskResponse{Id: task.ID, Status: constants.JOBFAILED}, err
	}
	return &pb.TaskResponse{Id: task.ID, Status: constants.JOBSUCCEED}, nil
}

func HandoutTasksForExecuting(task *constants.TaskCache) (string, int, error) {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		return task.ID, constants.JOBFAILED, fmt.Errorf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewTaskServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), constants.GRPC_TIMEOUT)
	defer cancel()

	payload := &pb.Payload{
		Format: int32(task.Payload.Format),
		Script: task.Payload.Script,
	}

	r, err := client.ExecuteTask(ctx, &pb.TaskRequest{
		Id:            task.ID,
		Payload:       payload,
		ExecutionTime: timestamppb.New(task.ExecutionTime),
		MaxRetryCount: int32(task.RetriesLeft),
	})

	if err != nil {
		// over time, no response
		return task.ID, constants.JOBFAILED, fmt.Errorf("could not execute update: %v", err)
	}
	log.Printf("Task %s executed with status %d", r.Id, r.Status)
	return r.Id, int(r.Status), nil
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

func RenewLease(taskID string, newLeaseTime time.Time) (bool, error) {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewLeaseServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), constants.GRPC_TIMEOUT)
	defer cancel()

	success, err := client.RenewLease(ctx, &pb.RenewLeaseRequest{
		Id:           taskID,
		NewLeaseTime: timestamppb.New(newLeaseTime),
	})

	if err != nil {
		return false, fmt.Errorf("could not renew lease: %v", err)
	}

	return success.Success, nil
}

func (s *ServerImpl) StopTask(ctx context.Context, in *pb.TaskStopRequest) (*pb.TaskResponse, error) {
	// Similar with ExecuteTask, but it is to stop a update
	return nil, nil // TODO
}

func StopRunningTask(runningTask *constants.RunTimeTask) (string, int, error) {
	// TODO
	return "", 0, nil
}
