package generator

import (
	"context"
	redis2 "git.woa.com/robingowang/MoreFun_SuperNova/pkg/database/redis"
	"git.woa.com/robingowang/MoreFun_SuperNova/utils/constants"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
)

func TestGenerateTask(t *testing.T) {
	mr, err := miniredis.Run()

	if err != nil {
		panic(err)
	}
	defer mr.Close()

	redis2.InitializeRedisRing(map[string]string{"shard1": mr.Addr()})

	name := "Test Task"
	taskType := 1
	schedule := "* * * * *"
	importance := 0.5
	payload := "test payload"
	callbackURL := "http://example.com/callback"

	task := GenerateTask(name, taskType, schedule, importance, payload, callbackURL)

	if task.Name != name || task.Type != taskType || task.Schedule != schedule ||
		task.Importance != importance || task.Payload != payload || task.CallbackURL != callbackURL {
		t.Errorf("GenerateTask() failed, expected task with name: %s, "+
			"type: %d, schedule: %s, importance %f, payload: %s, callbackURL: %s, got: %+v",
			name, taskType, schedule, importance, payload, callbackURL, task)
	}
}

func TestMarshallTask(t *testing.T) {
	name := "Test Task"
	taskType := 1
	schedule := "* * * * *"
	importance := 0.5
	payload := "test payload"
	callbackURL := "http://example.com/callback"

	task := GenerateTask(name, taskType, schedule, importance, payload, callbackURL)
	taskJson := marshallTask(task)

	if len(taskJson) == 0 {
		t.Errorf("marshallTask() failed, expected taskJson with length > 0, got: %d",
			len(taskJson))
	}
}

func TestStoreDataToRedis(t *testing.T) {
	mr, err := miniredis.Run()

	if err != nil {
		panic(err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	redis2.InitializeRedisRing(map[string]string{"shard1": mr.Addr()})

	name := "Test Task"
	taskType := 1
	schedule := "* * * * *"
	importance := 0.5
	payload := "test payload"
	callbackURL := "http://example.com/callback"

	task := GenerateTask(name, taskType, schedule, importance, payload, callbackURL)
	taskJson := marshallTask(task)

	assert.NoError(t, err)

	val, err := client.Get(context.Background(), task.ID).Result()
	assert.NoError(t, err)
	assert.Equal(t, string(taskJson), val)
}

func TestStoreDataToMongoDB(t *testing.T) {
	name := "Test Task"
	taskType := 1
	schedule := "* * * * *"
	importance := 0.5
	payload := "test payload"
	callbackURL := "http://example.com/callback"

	task := GenerateTask(name, taskType, schedule, importance, payload, callbackURL)

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	assert.NoError(t, err)

	collection := client.Database("test").Collection("tasks")

	_, err = collection.InsertOne(context.Background(), task)
	assert.NoError(t, err)

	var result map[string]interface{}
	err = collection.FindOne(context.Background(), bson.M{"_id": task.ID}).Decode(&result)
	assert.NoError(t, err)

	assert.Equal(t, task.ID, result["_id"])
	assert.Equal(t, task.Name, result["name"])
}

func TestSetupMongoConnection(t *testing.T) {
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
	assert.NoError(t, err)

	assert.NotNil(t, client)
}
