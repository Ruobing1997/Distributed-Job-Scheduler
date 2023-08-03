package mongoDB

import (
	"context"
	"fmt"
	"git.woa.com/robingowang/MoreFun_SuperNova/utils/constants"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupMongoConnection(client *mongo.Client) {
	// Send a ping to confirm a successful connection
	var result bson.M
	if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Decode(&result); err != nil {
		panic(err)
	}
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")
}

func StoreDataToMongoDB(task *constants.TaskDB) {
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
