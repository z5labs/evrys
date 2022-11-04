package eventstore

import (
	"context"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func TestMongo(t *testing.T) {
	// setup
	req := require.New(t)
	ctx := context.Background()

	// container init
	contReq := testcontainers.ContainerRequest{
		Image: "mongo:6.0.2",
		Env: map[string]string{
			"MONGO_INITDB_ROOT_USERNAME": "root",
			"MONGO_INITDB_ROOT_PASSWORD": "example",
		},
		ExposedPorts: []string{"27017:27017"},
		WaitingFor:   wait.ForLog("Waiting for connections"),
	}
	mongoC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: contReq,
		Started:          true,
	})
	req.NoError(err, "failed to create mongo container")
	defer mongoC.Terminate(ctx)

	// mongo verification
	uri := "mongodb://root:example@localhost:27017"
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	req.NoError(err, "failed to connect to mongo")

	req.NoError(client.Ping(ctx, readpref.Primary()), "failed to ping mongo")

	// impl setup
	db := "testdb"
	collName := "testcoll"
	config := &MongoConfig{
		Host:       "localhost",
		Port:       "27017",
		Username:   "root",
		Password:   "example",
		Database:   db,
		Collection: collName,
	}

	mongoImpl, err := NewMongoEventStoreImpl(ctx, config)
	req.NoError(err, "failed to create mongo event store")

	// data setup
	_event := event.New()
	id := "some_random_id"
	_event.SetID(id)
	_event.SetSubject("test")
	_event.SetSource("mongo_test")
	_event.SetTime(time.Now())
	_event.SetSpecVersion(event.CloudEventsVersionV1)
	_event.SetType("test")
	_event.SetData(*event.StringOfApplicationJSON(), map[string]interface{}{"hello": "world"})

	// actual test
	req.NoError(mongoImpl.PutEvent(ctx, &_event), "failed to put event")

	coll := client.Database(db).Collection(collName)
	filter := bson.D{{Key: "id", Value: id}}
	cursor, err := coll.Find(ctx, filter)
	req.NoError(err, "failed to find record")

	var results []bson.D
	req.NoError(cursor.All(ctx, &results), "failed to convert to array")

	req.Len(results, 1, "returned slice not of correct length")
}
