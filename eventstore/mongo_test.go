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
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func TestMongoConfig_Validate(t *testing.T) {
	req := require.New(t)

	t.Run("invalid config - no host", func(t *testing.T) {
		conf := MongoConfig{
			Port:       "1234",
			Username:   "username",
			Password:   "dfasdfad",
			Database:   "dfasdfas",
			Collection: "dfads",
		}
		var valError *ValidationError
		req.ErrorAs(conf.Validate(), &valError, "config should not have validated")
	})

	t.Run("invalid config - port none", func(t *testing.T) {
		conf := MongoConfig{
			Host:       "dasdfas",
			Username:   "username",
			Password:   "dfasdfad",
			Database:   "dfasdfas",
			Collection: "dfads",
		}
		var valError *ValidationError
		req.ErrorAs(conf.Validate(), &valError, "config should not have validated")
	})

	t.Run("invalid config - port invalid", func(t *testing.T) {
		conf := MongoConfig{
			Host:       "dasdfas",
			Port:       "1234cdads5",
			Username:   "username",
			Password:   "dfasdfad",
			Database:   "dfasdfas",
			Collection: "dfads",
		}
		var valError *ValidationError
		req.ErrorAs(conf.Validate(), &valError, "config should not have validated")
	})

	t.Run("invalid config - no username", func(t *testing.T) {
		conf := MongoConfig{
			Host:       "dasdfas",
			Port:       "1234",
			Password:   "dfasdfad",
			Database:   "dfasdfas",
			Collection: "dfads",
		}
		var valError *ValidationError
		req.ErrorAs(conf.Validate(), &valError, "config should not have validated")
	})

	t.Run("invalid config - no password", func(t *testing.T) {
		conf := MongoConfig{
			Host:       "dasdfas",
			Port:       "1234",
			Username:   "username",
			Database:   "dfasdfas",
			Collection: "dfads",
		}
		var valError *ValidationError
		req.ErrorAs(conf.Validate(), &valError, "config should not have validated")
	})

	t.Run("invalid config - no database", func(t *testing.T) {
		conf := MongoConfig{
			Host:       "dasdfas",
			Port:       "1234",
			Username:   "username",
			Password:   "dfasdfad",
			Collection: "dfads",
		}
		var valError *ValidationError
		req.ErrorAs(conf.Validate(), &valError, "config should not have validated")
	})

	t.Run("invalid config - no collection", func(t *testing.T) {
		conf := MongoConfig{
			Host:     "dasdfas",
			Port:     "1234",
			Username: "username",
			Password: "dfasdfad",
			Database: "dfasdfas",
		}
		var valError *ValidationError
		req.ErrorAs(conf.Validate(), &valError, "config should not have validated")
	})

	t.Run("valid config", func(t *testing.T) {
		conf := MongoConfig{
			Host:       "something",
			Port:       "1234",
			Username:   "username",
			Password:   "dfasdfad",
			Database:   "dfasdfas",
			Collection: "dfads",
		}
		req.NoError(conf.Validate(), "config should have validated")
	})
}

func TestNewMongoEventStoreImpl(t *testing.T) {
	req := require.New(t)
	t.Run("nil ctx", func(t *testing.T) {
		_, err := NewMongoEventStoreImpl(nil, nil)
		req.ErrorContains(err, "context can not be nil", "error is not target error")
	})

	t.Run("nil config", func(t *testing.T) {
		_, err := NewMongoEventStoreImpl(context.TODO(), nil)
		req.ErrorContains(err, "config can not be nil", "error is not target error")
	})

	t.Run("invalid config", func(t *testing.T) {
		conf := MongoConfig{
			Host:       "something",
			Port:       "1234",
			Username:   "username",
			Database:   "dfasdfas",
			Collection: "dfads",
		}
		_, err := NewMongoEventStoreImpl(context.TODO(), &conf)
		var valError *ValidationError
		req.ErrorAs(err, &valError, "expected validation error")
	})

	t.Run("mongo connection error", func(t *testing.T) {
		conf := MongoConfig{
			Host:       "localhost",
			Port:       "27017",
			Password:   "asdfasdf",
			Username:   "username",
			Database:   "testdb",
			Collection: "testcoll",
		}
		_, err := NewMongoEventStoreImpl(context.TODO(), &conf)
		var connErr *ConnectionError
		req.ErrorAs(err, &connErr, "expected connection error")
	})
}

func TestMongoIntegration(t *testing.T) {
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

	err = client.Ping(ctx, readpref.Primary())
	req.NoError(err, "failed to ping mongo")

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
	curTime := time.Now().UTC()
	_event.SetID(id)
	_event.SetSubject("test")
	_event.SetSource("mongo_test")
	_event.SetTime(curTime)
	_event.SetSpecVersion(event.CloudEventsVersionV1)
	_event.SetType("test")
	_event.SetData(*event.StringOfApplicationJSON(), map[string]interface{}{"hello": "world"})

	// actual test
	err = mongoImpl.PutEvent(ctx, &_event)
	req.NoError(err, "failed to put event")

	coll := client.Database(db).Collection(collName)
	filter := bson.D{{Key: "id", Value: id}}
	cursor, err := coll.Find(ctx, filter)
	req.NoError(err, "failed to find record")

	var results []bson.D
	err = cursor.All(ctx, &results)
	req.NoError(err, "failed to convert to array")

	req.Len(results, 1, "returned slice not of correct length")

	result := results[0]
	var idv, spec, source, _type, sub, dct, tm, dt = false, false, false, false, false, false, false, false
	for key, value := range result.Map() {
		switch key {
		case "id":
			req.Equal(id, value, "id not expected value")
			idv = true
			continue
		case "specversion":
			req.Equal("1.0", value, "spec version not expected value")
			spec = true
			continue
		case "source":
			req.Equal("mongo_test", value, "source not expected value")
			source = true
			continue
		case "type":
			req.Equal("test", value, "type not expected value")
			_type = true
			continue
		case "subject":
			req.Equal("test", value, "subject not expected value")
			sub = true
			continue
		case "datacontenttype":
			req.Equal("application/json", value, "datacontenttype not expected value")
			dct = true
			continue
		case "time":
			req.Equal(curTime.Format(time.RFC3339Nano), value, "time is not equal")
			tm = true
			continue
		case "data":
			d := value.(primitive.D)
			v := d[0]
			req.Equal("hello", v.Key, "key not expected value")
			req.Equal("world", v.Value, "value not of expected value")
			dt = true
			continue
		}
	}
	req.True(idv && spec && source && _type && sub && dct && tm && dt, "all values have not been verified")
}
