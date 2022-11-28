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

package notification

import (
	"context"
	"errors"

	evryspb "github.com/z5labs/evrys/proto"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

var kafkaBus = "kafka"

// KafkaBus
type KafkaBus struct {
	kafka *kafka.Writer

	log *zap.Logger
}

// KafkaConfig
type KafkaConfig struct {
	Addresses              []string
	AllowAutoTopicCreation bool
	Logger                 *zap.Logger
}

var (
	ErrMissingKafkaAddrs = errors.New("kafka addresses are required")
)

// NewKafkaBus
func NewKafkaBus(cfg KafkaConfig) (*KafkaBus, error) {
	err := validateKafkaConfig(cfg)
	if err != nil {
		return nil, ConfigurationError{
			Bus:   kafkaBus,
			Cause: err,
		}
	}

	bus := &KafkaBus{
		kafka: &kafka.Writer{
			Addr:                   kafka.TCP(cfg.Addresses...),
			AllowAutoTopicCreation: cfg.AllowAutoTopicCreation,
		},
		log: cfg.Logger,
	}

	return bus, nil
}

func validateKafkaConfig(cfg KafkaConfig) error {
	validators := []func(KafkaConfig) error{
		validateKafkaLogger,
		validateKafkaAddrs,
	}
	for _, validator := range validators {
		err := validator(cfg)
		if err != nil {
			return err
		}
	}
	return nil
}

func validateKafkaLogger(cfg KafkaConfig) error {
	if cfg.Logger == nil {
		return ErrMissingLogger
	}
	return nil
}

func validateKafkaAddrs(cfg KafkaConfig) error {
	if cfg.Addresses == nil || len(cfg.Addresses) == 0 {
		return ErrMissingKafkaAddrs
	}
	return nil
}

// Publish
func (bus *KafkaBus) Publish(ctx context.Context, n *evryspb.Notification) error {
	err := validateNotification(n)
	if err != nil {
		bus.log.Error(
			"given an invalid notification for publishing",
			zap.String("event_source", n.EventSource),
			zap.String("event_id", n.EventId),
			zap.String("event_type", n.EventType),
			zap.Error(err),
		)
		return ValidationError{
			Cause: err,
		}
	}

	b, err := proto.Marshal(n)
	if err != nil {
		return MarshalError{
			Protocol: "protobuf",
			Cause:    err,
		}
	}

	msg := kafka.Message{
		Topic: n.EventType,
		Value: b,
	}
	bus.log.Info(
		"publishing event notification",
		zap.String("event_source", n.EventSource),
		zap.String("event_id", n.EventId),
		zap.String("event_type", n.EventType),
	)
	err = bus.kafka.WriteMessages(ctx, msg)
	if err != nil {
		bus.log.Error(
			"failed to publish event notification",
			zap.String("event_source", n.EventSource),
			zap.String("event_id", n.EventId),
			zap.String("event_type", n.EventType),
			zap.Error(err),
		)
		return Error{
			Bus:   kafkaBus,
			Cause: err,
		}
	}
	bus.log.Info(
		"published event notification",
		zap.String("event_source", n.EventSource),
		zap.String("event_id", n.EventId),
		zap.String("event_type", n.EventType),
	)
	return nil
}
