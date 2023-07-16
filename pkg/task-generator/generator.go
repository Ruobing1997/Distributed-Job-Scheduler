package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"git.woa.com/robingowang/MoreFun_SuperNova/pkg/database/redis"
	"git.woa.com/robingowang/MoreFun_SuperNova/utils/constants"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GenerateTask(name string, taskType int, schedule string, importance float64,
	payload string, callbackURL string) constants.Task {
	id := uuid.New().String()
	task := constants.Task{
		ID:          id,
		Name:        name,
		Type:        taskType,
		Schedule:    schedule,
		Importance:  importance,
		Payload:     payload,
		CallbackURL: callbackURL,
		Status:      0,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
		Retries:     0,
		Result:      "",
	}

	return task
}

func marshallTask(task constants.Task) []byte {
	taskJson, err := json.Marshal(task)
	if err != nil {
		log.Fatal(err)
	}
	return taskJson
}

// StoreDataToRedis use redis ringOptions to store data, the ringOptions uses Consistent Hashing
func StoreDataToRedis(key string, value []byte) error {
	redisRing := redis.GetRedisRing()

	err := redisRing.Set(context.Background(), key, value, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to set key %s: %v", key, err)
	}

	fmt.Printf("Stored key %s with value %s to the Ring\n", key, value)
	return nil
}

func StoreDataToMongoDB(task *constants.Task) {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(constants.MONGOURI).SetServerAPIOptions(serverAPI)
	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	setupMongoConnection(client)
	result, err := client.Database("job_scheduler_test").Collection("tasks").InsertOne(context.Background(), task)
	if err != nil {
		fmt.Println("InsertOne() result ERROR:", err)
	} else {
		fmt.Println("InsertOne() result:", result.InsertedID)
	}
}

func setupMongoConnection(client *mongo.Client) {
	// Send a ping to confirm a successful connection
	var result bson.M
	if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Decode(&result); err != nil {
		panic(err)
	}
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")
}
