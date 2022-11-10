package eventstore

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/cloudevents/sdk-go/v2/event"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
)

// MongoConfig defines the configuration to connect to mongodb
type MongoConfig struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`

	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`

	Database   string `mapstructure:"database"`
	Collection string `mapstructure:"collection"`
}

// Validate ensures mongo config is correct
func (m *MongoConfig) Validate() error {
	if m.Host == "" {
		return NewValidationError("MongoConfig", "Host", "Empty string")
	}

	if m.Port == "" {
		return NewValidationError("MongoConfig", "Port", "Empty string")
	} else {
		_, err := strconv.Atoi(m.Port)
		if err != nil {
			return NewValidationError("MongoConfig", "Port", "not a number")
		}
	}

	if m.Username == "" {
		return NewValidationError("MongoConfig", "Username", "Empty string")
	}

	if m.Password == "" {
		return NewValidationError("MongoConfig", "Password", "Empty string")
	}

	if m.Database == "" {
		return NewValidationError("MongoConfig", "Database", "Empty string")
	}

	if m.Collection == "" {
		return NewValidationError("MongoConfig", "Collection", "Empty string")
	}

	return nil
}

func (m *MongoConfig) getURI() string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%s", m.Username, m.Password, m.Host, m.Port)
}

// MongoEventStoreImpl is the event store implementation for mongodb
type MongoEventStoreImpl struct {
	config MongoConfig
	logger *zap.Logger
	client *mongo.Client
}

// NewMongoEventStoreImpl constructs and initializes a *MongoEventStoreImpl
func NewMongoEventStoreImpl(ctx context.Context, _config MongoConfig) (*MongoEventStoreImpl, error) {
	if ctx == nil {
		return nil, errors.New("context can not be nil")
	}

	err := _config.Validate()
	if err != nil {
		return nil, fmt.Errorf("invalid config, %w", err)
	}

	impl := &MongoEventStoreImpl{
		config: _config,
		logger: zap.L().With(zap.String("source", "MongoEventStoreImpl")),
	}

	err = impl.init(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to init, %w", err)
	}

	return impl, nil
}

func (m *MongoEventStoreImpl) init(ctx context.Context) error {
	m.logger.Debug("attempting to open connection to mongo")
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(m.config.getURI()))
	if err != nil {
		m.logger.Error("failed to connect to mongo", zap.Error(err))
		return NewConnectionError("mongo", err)
	}
	m.logger.Debug("successfully connected to mongo")

	// attempt ping
	m.logger.Debug("attempting ping")
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		m.logger.Error("failed to ping mongo", zap.Error(err))
		return NewConnectionError("mongo", err)
	}
	m.logger.Debug("successfully pinged mongo")

	m.client = client
	return nil
}

// PutEvent puts an event into mongo and implements the interface PutEvent
func (m *MongoEventStoreImpl) PutEvent(ctx context.Context, event *event.Event) error {
	coll := m.client.Database(m.config.Database).Collection(m.config.Collection)

	m.logger.Debug("attempting to marshal event to json")
	raw, err := event.MarshalJSON()
	if err != nil {
		m.logger.Error("failed to marshal event to json", zap.Error(err))
		return NewMarshalError("*event.Event", "json", err)
	}
	m.logger.Debug("successfully marshaled event to json")

	m.logger.Debug("attempting to marshal json to bson")
	var bdoc interface{}
	err = bson.UnmarshalExtJSON(raw, true, &bdoc)
	if err != nil {
		m.logger.Error("failed to marshal json to bson", zap.Error(err))
		return NewMarshalError("json", "bson", err)
	}
	m.logger.Debug("successfully marshaled json to bson")

	m.logger.Debug("attempting to insert event")
	_, err = coll.InsertOne(ctx, bdoc)
	if err != nil {
		m.logger.Error("failed to insert event", zap.Error(err))
		return NewPutError("mongo", "event", err)
	}
	m.logger.Debug("successfully inserted event")

	return nil
}
