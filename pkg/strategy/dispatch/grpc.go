package dispatch

import (
	"context"
	"fmt"
	pb "git.woa.com/robingowang/MoreFun_SuperNova/pkg/strategy/dispatch/proto"
	task_executor "git.woa.com/robingowang/MoreFun_SuperNova/pkg/task-executor"
	generator "git.woa.com/robingowang/MoreFun_SuperNova/pkg/task-generator"
	task_manager "git.woa.com/robingowang/MoreFun_SuperNova/pkg/task-manager"
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
		return task.ID, constants.JOBFAILED, fmt.Errorf("could not execute task: %v", err)
	}
	log.Printf("Task %s executed with status %d", r.Id, r.Status)
	return r.Id, int(r.Status), nil
}

func (s *ServerImpl) RenewLease(ctx context.Context, in *pb.RenewLeaseRequest) (*pb.RenewLeaseResponse, error) {
	// worker should ask for a new lease before the current lease expires
	// if the worker fails to do so, the task will be handed out to another worker depends on retry count
	// if the worker fails to do so and the retry count is 0, the task will be marked as failed
	currentLease := task_manager.GetLeaseByID(in.Id)
	if time.Now().After(currentLease) {
		return &pb.RenewLeaseResponse{Success: false},
			fmt.Errorf("lease for task %s has expired", in.Id)
	}

	newLease := time.Now().Add(constants.LEASE_DURATION)
	task_manager.UpdateLeaseByID(in.Id, newLease)
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
