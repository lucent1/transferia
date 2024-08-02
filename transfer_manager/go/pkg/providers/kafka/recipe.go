package kafka

import (
	"context"
	"os"
	"strings"

	"github.com/doublecloud/tross/cloud/dataplatform/testcontainer/kafka"
	"github.com/doublecloud/tross/library/go/core/xerrors"
	server "github.com/doublecloud/tross/transfer_manager/go/pkg/abstract/model"
	"github.com/doublecloud/tross/transfer_manager/go/pkg/providers/logbroker"
)

func ContainerNeeded() bool {
	return os.Getenv("USE_TESTCONTAINERS") == "1" && os.Getenv("KAFKA_RECIPE_BROKER_LIST") == ""
}

func SourceRecipe() (*KafkaSource, error) {
	if ContainerNeeded() {
		if err := StartKafkaContainer(); err != nil {
			return nil, xerrors.Errorf("unable to setup container: %w", err)
		}
	}
	brokers := os.Getenv("KAFKA_RECIPE_BROKER_LIST")
	src := new(KafkaSource)
	src.Connection = &KafkaConnectionOptions{
		ClusterID:    "",
		TLS:          logbroker.DisabledTLS,
		TLSFile:      "",
		Brokers:      []string{brokers},
		SubNetworkID: "",
	}
	src.Auth = &KafkaAuth{
		Enabled:   false,
		Mechanism: "",
		User:      "",
		Password:  "",
	}
	src.BufferSize = server.BytesSize(1024)
	return src, nil
}

func MustSourceRecipe() *KafkaSource {
	result, err := SourceRecipe()
	if err != nil {
		panic(err)
	}
	return result
}

func StartKafkaContainer() error {
	cntr, err := kafka.RunContainer(context.Background())
	if err != nil {
		return xerrors.Errorf("unable to start kafka container: %w", err)
	}
	brokers, err := cntr.Brokers(context.Background())
	if err != nil {
		return xerrors.Errorf("unable fetch brokers: %w", err)
	}
	if err := os.Setenv("KAFKA_RECIPE_BROKER_LIST", strings.Join(brokers, ",")); err != nil {
		return xerrors.Errorf("unable to set broker list env: %w", err)
	}
	return nil
}

func DestinationRecipe() (*KafkaDestination, error) {
	if ContainerNeeded() {
		if err := StartKafkaContainer(); err != nil {
			return nil, xerrors.Errorf("unable to setup container: %w", err)
		}
	}
	brokers := os.Getenv("KAFKA_RECIPE_BROKER_LIST")

	dst := new(KafkaDestination)
	dst.Connection = &KafkaConnectionOptions{
		ClusterID:    "",
		TLS:          logbroker.DisabledTLS,
		TLSFile:      "",
		Brokers:      []string{brokers},
		SubNetworkID: "",
	}
	dst.Auth = &KafkaAuth{
		Enabled:   false,
		Mechanism: "",
		User:      "",
		Password:  "",
	}
	dst.FormatSettings = server.SerializationFormat{
		Name:             server.SerializationFormatAuto,
		Settings:         nil,
		SettingsKV:       nil,
		BatchingSettings: nil,
	}
	return dst, nil
}
