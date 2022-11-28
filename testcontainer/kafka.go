// Copyright 2022 Z5Labs and Contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package testcontainer

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
)

// KafkaCluster
type KafkaCluster struct {
	cfg kafkaOptions

	zookeeper testcontainers.Container
	kafka     testcontainers.Container
}

type kafkaOptions struct {
	tag                string
	clusterNetworkName string
	zookeeperPort      string
	kafkaBrokerPort    string
	kafkaClientPort    string
}

func (ko kafkaOptions) zookeeperImage() string {
	return fmt.Sprintf("confluentinc/cp-zookeeper:%s", ko.tag)
}

func (ko kafkaOptions) kafkaImage() string {
	return fmt.Sprintf("confluentinc/cp-kafka:%s", ko.tag)
}

// KafkaOption
type KafkaOption func(*kafkaOptions)

// WithKafkaTag
func WithKafkaTag(tag string) KafkaOption {
	return func(ko *kafkaOptions) {
		ko.tag = tag
	}
}

// WithClusterNetworkName
func WithClusterNetworkName(name string) KafkaOption {
	return func(ko *kafkaOptions) {
		ko.clusterNetworkName = name
	}
}

// WithZooKeeperPort
func WithZooKeeperPort(port string) KafkaOption {
	return func(ko *kafkaOptions) {
		ko.zookeeperPort = port
	}
}

// WithKafkaBrokerPort
func WithKafkaBrokerPort(port string) KafkaOption {
	return func(ko *kafkaOptions) {
		ko.kafkaBrokerPort = port
	}
}

// WithKafkaClientPort
func WithKafkaClientPort(port string) KafkaOption {
	return func(ko *kafkaOptions) {
		ko.kafkaClientPort = port
	}
}

// NewKafkaCluster
func NewKafkaCluster(ctx context.Context, opts ...KafkaOption) (*KafkaCluster, error) {
	kopts := kafkaOptions{
		tag:                "latest",
		clusterNetworkName: "kafka-cluster",
		zookeeperPort:      "2181",
		kafkaBrokerPort:    "9092",
		kafkaClientPort:    "9093",
	}
	for _, opt := range opts {
		opt(&kopts)
	}
	err := validateKafkaClusterConfig(kopts)
	if err != nil {
		return nil, ValidationError{
			Cause: err,
		}
	}

	// creates a network, so kafka and zookeeper can communicate directly
	network, err := testcontainers.GenericNetwork(ctx, testcontainers.GenericNetworkRequest{
		NetworkRequest: testcontainers.NetworkRequest{Name: kopts.clusterNetworkName},
	})
	if err != nil {
		return nil, FailedToCreateNetwork{
			Name:  kopts.clusterNetworkName,
			Cause: err,
		}
	}
	dockerNetwork := network.(*testcontainers.DockerNetwork)

	// Create zookeeper container
	zookeeper, err := createZookeeperContainer(ctx, dockerNetwork, kopts)
	if err != nil {
		return nil, FailedToCreateContainer{
			Image: kopts.zookeeperImage(),
			Cause: err,
		}
	}

	// Create kafka container
	kafka, err := createKafkaContainer(ctx, dockerNetwork, kopts)
	if err != nil {
		return nil, FailedToCreateContainer{
			Image: kopts.kafkaImage(),
			Cause: err,
		}
	}

	kc := &KafkaCluster{
		cfg:       kopts,
		zookeeper: zookeeper,
		kafka:     kafka,
	}
	return kc, nil
}

// Start
func (kc *KafkaCluster) Start(ctx context.Context) error {
	err := kc.zookeeper.Start(ctx)
	if err != nil {
		return FailedToStartContainer{
			Image: kc.cfg.zookeeperImage(),
			Cause: err,
		}
	}
	err = kc.kafka.Start(ctx)
	if err != nil {
		return FailedToStartContainer{
			Image: kc.cfg.kafkaImage(),
			Cause: err,
		}
	}

	kafkaStartFile, err := ioutil.TempFile("", "testcontainers_start.sh")
	if err != nil {
		return err
	}
	defer os.Remove(kafkaStartFile.Name())

	// needs to set KAFKA_ADVERTISED_LISTENERS with the exposed kafka port
	exposedHost, err := kc.GetKafkaHost(ctx)
	if err != nil {
		return err
	}

	kafkaStartFile.WriteString("#!/bin/bash \n")
	kafkaStartFile.WriteString("export KAFKA_ADVERTISED_LISTENERS='PLAINTEXT://" + exposedHost + ",BROKER://kafka:" + kc.cfg.kafkaBrokerPort + "'\n")
	kafkaStartFile.WriteString(". /etc/confluent/docker/bash-config \n")
	kafkaStartFile.WriteString("/etc/confluent/docker/configure \n")
	kafkaStartFile.WriteString("/etc/confluent/docker/launch \n")

	err = kc.kafka.CopyFileToContainer(ctx, kafkaStartFile.Name(), "testcontainers_start.sh", 0700)
	if err != nil {
		return err
	}
	return nil
}

type KafkaClusterTerminationError struct {
	KafkaContainerErr     error
	ZookeeperContainerErr error
}

func (e KafkaClusterTerminationError) Error() string {
	return fmt.Sprintf(
		"failed to terminate kafka cluster containers: \n\tkafka container: %s\n\tzookeeper container: %s",
		e.KafkaContainerErr,
		e.ZookeeperContainerErr,
	)
}

func (kc *KafkaCluster) Terminate(ctx context.Context) error {
	kafkaErr := kc.kafka.Terminate(ctx)
	zookeeperErr := kc.zookeeper.Terminate(ctx)
	if kafkaErr == nil && zookeeperErr == nil {
		return nil
	}
	return KafkaClusterTerminationError{
		KafkaContainerErr:     kafkaErr,
		ZookeeperContainerErr: zookeeperErr,
	}
}

// GetKafkaHost
func (kc *KafkaCluster) GetKafkaHost(ctx context.Context) (string, error) {
	host, err := kc.kafka.Host(ctx)
	if err != nil {
		return "", err
	}
	port, err := kc.kafka.MappedPort(ctx, nat.Port(kc.cfg.kafkaClientPort))
	if err != nil {
		return "", err
	}
	return host + ":" + port.Port(), nil
}

func validateKafkaClusterConfig(kopts kafkaOptions) error {
	validators := []func(kafkaOptions) error{
		validateKafkaTag,
		validateKafkaClusterNetworkName,
		validateZookeeperPort,
		validateKafkaBrokerPort,
		validateKafkaClientPort,
	}
	for _, validator := range validators {
		err := validator(kopts)
		if err != nil {
			return err
		}
	}
	return nil
}

var (
	ErrInvalidKafkaTag         = errors.New("invalid kafka docker tag")
	ErrInvalidKafkaNetworkName = errors.New("invalid kafka cluster network name")
	ErrInvalidZookeeperPort    = errors.New("invalid kafka broker port")
	ErrInvalidKafkaBrokerPort  = errors.New("invalid kafka broker port")
	ErrInvalidKafkaClientPort  = errors.New("invalid kafka broker port")
)

func validateKafkaTag(kopts kafkaOptions) error {
	tag := strings.TrimSpace(kopts.tag)
	if tag == "latest" {
		return nil
	}
	v := strings.Split(tag, ",")
	if len(v) != 3 {
		return ErrInvalidKafkaTag
	}
	for _, s := range v {
		_, err := strconv.Atoi(s)
		if err != nil {
			return ErrInvalidKafkaTag
		}
	}
	return nil
}

func validateKafkaClusterNetworkName(kopts kafkaOptions) error {
	name := strings.TrimSpace(kopts.clusterNetworkName)
	if len(name) == 0 {
		return ErrInvalidKafkaNetworkName
	}
	return nil
}

func validateZookeeperPort(kopts kafkaOptions) error {
	_, err := strconv.Atoi(kopts.zookeeperPort)
	if err != nil {
		return ErrInvalidZookeeperPort
	}
	return nil
}

func validateKafkaBrokerPort(kopts kafkaOptions) error {
	_, err := strconv.Atoi(kopts.kafkaBrokerPort)
	if err != nil {
		return ErrInvalidKafkaBrokerPort
	}
	return nil
}

func validateKafkaClientPort(kopts kafkaOptions) error {
	_, err := strconv.Atoi(kopts.kafkaClientPort)
	if err != nil {
		return ErrInvalidKafkaClientPort
	}
	return nil
}

func createZookeeperContainer(ctx context.Context, network *testcontainers.DockerNetwork, kopts kafkaOptions) (testcontainers.Container, error) {
	zookeeperContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:          kopts.zookeeperImage(),
			ExposedPorts:   []string{kopts.zookeeperPort},
			Env:            map[string]string{"ZOOKEEPER_CLIENT_PORT": kopts.zookeeperPort, "ZOOKEEPER_TICK_TIME": "2000"},
			Networks:       []string{network.Name},
			NetworkAliases: map[string][]string{network.Name: {"zookeeper"}},
		},
	})
	if err != nil {
		return nil, err
	}
	return zookeeperContainer, nil
}

func createKafkaContainer(ctx context.Context, network *testcontainers.DockerNetwork, kopts kafkaOptions) (testcontainers.Container, error) {
	kafkaContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        kopts.kafkaImage(),
			ExposedPorts: []string{kopts.kafkaClientPort},
			Env: map[string]string{
				"KAFKA_BROKER_ID":                        "1",
				"KAFKA_ZOOKEEPER_CONNECT":                "zookeeper:" + kopts.zookeeperPort,
				"KAFKA_LISTENERS":                        "PLAINTEXT://0.0.0.0:" + kopts.kafkaClientPort + ",BROKER://0.0.0.0:" + kopts.kafkaBrokerPort,
				"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP":   "BROKER:PLAINTEXT,PLAINTEXT:PLAINTEXT",
				"KAFKA_INTER_BROKER_LISTENER_NAME":       "BROKER",
				"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR": "1",
			},
			Networks:       []string{network.Name},
			NetworkAliases: map[string][]string{network.Name: {"kafka"}},
			// the container only starts when it finds and run /testcontainers_start.sh
			Cmd: []string{"sh", "-c", "while [ ! -f /testcontainers_start.sh ]; do sleep 0.1; done; /testcontainers_start.sh"},
		},
	})
	if err != nil {
		return nil, err
	}
	return kafkaContainer, nil
}
