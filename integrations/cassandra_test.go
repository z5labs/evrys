package integrations

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

/*
create keyspace example with replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };

	create table example.batches(pk int, ck int, description text, PRIMARY KEY(pk, ck));
	// INSERT INTO mytable JSON '{ "\"myKey\"": 0, "value": 0}';
*/
func TestCassandra(t *testing.T) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:      "cassandra:4.1",
		WaitingFor: wait.ForExposedPort(),
	}
	cassandraContainers, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Error(err)
	}
	defer cassandraContainers.Terminate(ctx)
}
