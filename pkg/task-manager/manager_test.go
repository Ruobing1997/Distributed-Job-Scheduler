package task_manager

import (
	"context"
	data_structure_redis "git.woa.com/robingowang/MoreFun_SuperNova/pkg/data-structure"
	"testing"
	"time"
)

func TestSubscribeToRedisChannel(t *testing.T) {
	// 初始化Redis客户端
	Init()
	Start()

	// 等待一小段时间，以确保SubscribeToRedisChannel已经开始订阅
	time.Sleep(2 * time.Second)

	// 发布消息
	err := redisClient.Publish(context.Background(),
		data_structure_redis.REDIS_CHANNEL, data_structure_redis.TASK_AVAILABLE).Err()
	if err != nil {
		t.Fatalf("Failed to publish message to Redis: %v", err)
	}

	// 等待消息或超时
	select {
	case receivedMessage := <-testMessageChannel:
		if receivedMessage != data_structure_redis.TASK_AVAILABLE {
			t.Errorf("Expected message '%s', but got '%s'", data_structure_redis.TASK_AVAILABLE, receivedMessage)
		}
	case <-time.After(20 * time.Second):
		t.Error("Did not receive expected message from Redis within timeout.")
	}
}
