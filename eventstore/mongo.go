package eventstore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// MongoConfig defines the configuration to connect to mongodb
type MongoConfig struct {
	Host string `mapstructure:"host" validate:"hostname,required"`
	Port string `mapstructure:"port" validate:"numeric,required"`

	Username string `mapstructure:"username" validate:"required"`
	Password string `mapstructure:"password" validate:"required"`

	Database   string `mapstructure:"database" validate:"required"`
	Collection string `mapstructure:"collection" validate:"required"`
}

// Validate ensures mongo config is correct
func (m *MongoConfig) Validate() error {
	return validator.New().Struct(m)
}

func (m *MongoConfig) getURI() string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%s", m.Username, m.Password, m.Host, m.Port)
}

// Mongo is the event store implementation for mongodb
type Mongo struct {
	config MongoConfig
	logger *zap.Logger
	client *mongo.Client
}

// NewMongo constructs and initializes a *Mongo
func NewMongo(ctx context.Context, config MongoConfig) (*Mongo, error) {
	if ctx == nil {
		return nil, errors.New("context can not be nil")
	}

	err := config.Validate()
	if err != nil {
		return nil, fmt.Errorf("invalid config, %w", err)
	}

	impl := &Mongo{
		config: config,
		logger: zap.L().With(zap.String("source", "MongoEventStoreImpl")),
	}

	err = impl.init(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to init, %w", err)
	}

	return impl, nil
}

func (m *Mongo) init(ctx context.Context) error {
	m.logger.Debug("attempting to open connection to mongo")
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(m.config.getURI()))
	if err != nil {
		m.logger.Error("failed to connect to mongo", zap.Error(err))
		return NewConnectionError("mongo", err)
	}
	m.logger.Debug("successfully connected to mongo")

	m.client = client
	return nil
}

// Append puts an event into mongo and implements the interface PutEvent
func (m *Mongo) Append(ctx context.Context, event *event.Event) error {
	coll := m.client.Database(m.config.Database).Collection(m.config.Collection)

	m.logger.Debug("attempting to marshal event to json",
		zap.String("event_id", event.ID()),
		zap.String("event_type", event.Type()),
		zap.String("event_source", event.Source()),
		zap.String("event_subject", event.Subject()),
	)
	raw, err := event.MarshalJSON()
	if err != nil {
		m.logger.Error("failed to marshal event to json",
			zap.Error(err),
			zap.String("event_id", event.ID()),
			zap.String("event_type", event.Type()),
			zap.String("event_source", event.Source()),
			zap.String("event_subject", event.Subject()),
		)
		return NewMarshalError("*event.Event", "json", err)
	}
	m.logger.Debug("successfully marshaled event to json",
		zap.String("event_id", event.ID()),
		zap.String("event_type", event.Type()),
		zap.String("event_source", event.Source()),
		zap.String("event_subject", event.Subject()),
	)

	m.logger.Debug("attempting to marshal json to bson",
		zap.String("event_id", event.ID()),
		zap.String("event_type", event.Type()),
		zap.String("event_source", event.Source()),
		zap.String("event_subject", event.Subject()),
	)
	var bdoc interface{}
	err = bson.UnmarshalExtJSON(raw, true, &bdoc)
	if err != nil {
		m.logger.Error("failed to marshal json to bson",
			zap.Error(err),
			zap.String("event_id", event.ID()),
			zap.String("event_type", event.Type()),
			zap.String("event_source", event.Source()),
			zap.String("event_subject", event.Subject()),
		)
		return NewMarshalError("json", "bson", err)
	}
	m.logger.Debug("successfully marshaled json to bson",
		zap.String("event_id", event.ID()),
		zap.String("event_type", event.Type()),
		zap.String("event_source", event.Source()),
		zap.String("event_subject", event.Subject()),
	)

	m.logger.Debug("attempting to insert event",
		zap.String("event_id", event.ID()),
		zap.String("event_type", event.Type()),
		zap.String("event_source", event.Source()),
		zap.String("event_subject", event.Subject()),
	)
	_, err = coll.InsertOne(ctx, bdoc)
	if err != nil {
		m.logger.Error("failed to insert event",
			zap.Error(err),
			zap.String("event_id", event.ID()),
			zap.String("event_type", event.Type()),
			zap.String("event_source", event.Source()),
			zap.String("event_subject", event.Subject()),
		)
		return NewPutError("mongo", "event", err)
	}
	m.logger.Info("successfully inserted event",
		zap.String("event_id", event.ID()),
		zap.String("event_type", event.Type()),
		zap.String("event_source", event.Source()),
		zap.String("event_subject", event.Subject()),
	)

	return nil
}

func (m *Mongo) GetByID(ctx context.Context, id string) (*event.Event, error) {
	m.logger.Info(
		"attempting to lookup event by id",
		zap.String("id", id),
	)

	coll := m.client.Database(m.config.Database).Collection(m.config.Collection)

	filter := bson.D{{Key: "id", Value: id}}
	result := coll.FindOne(ctx, filter)
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			m.logger.Error(
				"event with id does not exist",
				zap.String("id", id),
				zap.Error(result.Err()),
			)
			return nil, NewQueryError("mongo", "*event.Event", true, result.Err())
		}
		m.logger.Error(
			"failed to lookup event by id",
			zap.String("id", id),
			zap.Error(result.Err()),
		)
		return nil, NewQueryError("mongo", "*event.Event", false, result.Err())
	}

	var b bson.D
	err := result.Decode(&b)
	if err != nil {
		return nil, err
	}

	// remove mongo id field since cloud events doesn't like it
	// it will always be the first index
	b, err = removeElement(b, 0)
	if err != nil {
		return nil, err
	}

	jsb, err := bson.MarshalExtJSON(b, false, false)
	if err != nil {
		return nil, err
	}

	ev := event.New()
	err = ev.UnmarshalJSON(jsb)
	if err != nil {
		return nil, err
	}

	return &ev, nil
}

func (m *Mongo) EventsBefore(ctx context.Context, until time.Time) (events []*event.Event, cursor any, err error) {
	return nil, nil, nil
}

func removeElement[T any](s []T, i int) ([]T, error) {
	// s is [1,2,3,4,5,6], i is 2

	// perform bounds checking first to prevent a panic!
	if i >= len(s) || i < 0 {
		return nil, fmt.Errorf("index is out of range. index is %d with slice length %d", i, len(s))
	}

	// copy first element (1) to index `i`. At this point,
	// `s` will be [1,2,1,4,5,6]
	s[i] = s[0]
	// Remove the first element from the slice by truncating it
	// This way, `s` will now include all the elements from index 1
	// until the element at the last index
	return s[1:], nil
}
