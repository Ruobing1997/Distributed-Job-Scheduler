package dispatch

import (
	"context"
	"fmt"
	data_structure_redis "git.woa.com/robingowang/MoreFun_SuperNova/pkg/data-structure"
	pb "git.woa.com/robingowang/MoreFun_SuperNova/pkg/strategy/dispatch/proto"
	"github.com/google/uuid"
	"io"
	"log"
)

type ServerImpl struct {
	pb.UnimplementedTaskManagerServiceServer
}

func (s *ServerImpl) TaskStream(stream pb.TaskManagerService_TaskStreamServer) error {
	req, err := stream.Recv()
	if err != nil {
		return err
	}
	workerIdentity, ok := req.Request.(*pb.TaskStreamRequest_WorkerIdentity)
	if !ok {
		return fmt.Errorf("first request should be worker identity")
	}

	workerID := workerIdentity.WorkerIdentity.WorkerId
	log.Printf("Worker %s connected", workerID)

	streamID := uuid.New().String()
	err = data_structure_redis.AddStreamIDWorkerIDMapping(context.Background(), workerID, streamID)
	defer func(ctx context.Context, streamID string) {
		err := data_structure_redis.RemoveStreamIDWorkerIDMapping(ctx, streamID)
		if err != nil {
			log.Printf("Error when removing streamID-workerID mapping: %v", err)
		}
	}(context.Background(), streamID)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// turn off stream
			return nil
		}
		if err != nil {
			return err
		}

		switch r := req.Request.(type) {
		case *pb.TaskStreamRequest_TaskResponse:
			// handle task response
			log.Printf("Received task response from worker %s: %v", workerID, r.TaskResponse)
		}
	}
}

func (s *ServerImpl) fetchTaskForWorker() *pb.TaskRequest {

}

func RenewLease(in *pb.TaskStreamRequest_RenewLease) (*pb.TaskStreamResponse, error) {
	err := data_structure_redis.SetLeaseWithID(in.RenewLease.Id,
		in.RenewLease.LeaseDuration.AsDuration())
	if err != nil {
		return &pb.TaskStreamResponse{
			Response: &pb.TaskStreamResponse_RenewLeaseResponse{
				RenewLeaseResponse: &pb.RenewLeaseResponse{
					Success: false,
				},
			},
		}, fmt.Errorf("lease for update %s has expired", in.RenewLease.Id)
	}
	return &pb.TaskStreamResponse{
		Response: &pb.TaskStreamResponse_RenewLeaseResponse{
			RenewLeaseResponse: &pb.RenewLeaseResponse{
				Success: true,
			},
		},
	}, nil
}
