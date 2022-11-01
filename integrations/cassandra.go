package integrations

import (
	"context"
	"fmt"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/gocql/gocql"
)

// CassandraConfig defines information needed to connect to cassandra
type CassandraConfig struct {
	Hosts        []string `mapstrcture:"hosts"`
	ProtoVersion int      `mapstructure:"protoVersion"`
	KeySpace     string   `mapstructure:"keySpace"`
	Table        string   `mapstructure:"table"`
}

// set defaults for unset variables
func (c *CassandraConfig) setDefaults() {
	if c.ProtoVersion == 0 {
		c.ProtoVersion = 4
	}
}

// CassandraImpl implmenets event store interfaces for use with cassandra
type CassandraImpl struct {
	config  *CassandraConfig
	cluster *gocql.ClusterConfig
}

// NewCassandraImpl returns an instance of CassandraImpl with a configuration
func NewCassandraImpl(_config *CassandraConfig) *CassandraImpl {
	impl := &CassandraImpl{
		config:  _config,
		cluster: nil,
	}

	impl.init()

	return impl
}

func (c *CassandraImpl) init() {
	c.config.setDefaults()
	c.cluster = gocql.NewCluster(c.config.Hosts...)
	c.cluster.ProtoVersion = c.config.ProtoVersion
	c.cluster.Keyspace = c.config.KeySpace
}

// PutEvent puts an event in cassandra and implements the PutEvent interface
func (c *CassandraImpl) PutEvent(ctx context.Context, event event.Event) error {
	session, err := c.cluster.CreateSession()
	if err != nil {
		return fmt.Errorf("faied to create session, %w", err)
	}
	defer session.Close()

	b := session.NewBatch(gocql.UnloggedBatch).WithContext(ctx)
	bytes, err := event.MarshalJSON()
	if err != nil {
		return fmt.Errorf("unable to marshall event to json, %w", err)
	}

	b.Entries = append(b.Entries, gocql.BatchEntry{
		Stmt:       fmt.Sprintf("INSERT INTO %s JSON '%s'", c.config.Table, bytes),
		Idempotent: true,
	})

	err = session.ExecuteBatch(b)
	if err != nil {
		return fmt.Errorf("failed to execute batch, %w", err)
	}

	return nil
}
