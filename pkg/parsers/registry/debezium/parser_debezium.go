package debezium

import (
	"runtime"

	"github.com/transferia/transferia/library/go/core/xerrors"
	"github.com/transferia/transferia/pkg/parsers"
	debeziumengine "github.com/transferia/transferia/pkg/parsers/registry/debezium/engine"
	"github.com/transferia/transferia/pkg/schemaregistry/confluent"
	"github.com/transferia/transferia/pkg/stats"
	"go.ytsaurus.tech/library/go/core/log"
)

func NewParserDebezium(inWrapped interface{}, _ bool, logger log.Logger, _ *stats.SourceStats) (parsers.Parser, error) {
	var client *confluent.SchemaRegistryClient = nil
	var err error
	switch in := inWrapped.(type) {
	case *ParserConfigDebeziumLb:
		if in.SchemaRegistryURL == "" {
			break
		}
		client, err = confluent.NewSchemaRegistryClientWithTransport(in.SchemaRegistryURL, in.TLSFile, logger)
		if err != nil {
			return nil, xerrors.Errorf("Unable to create schema registry client: %w", err)
		}
		client.SetCredentials(in.Username, in.Password)
	case *ParserConfigDebeziumCommon:
		if in.SchemaRegistryURL == "" {
			break
		}
		client, err = confluent.NewSchemaRegistryClientWithTransport(in.SchemaRegistryURL, in.TLSFile, logger)
		if err != nil {
			return nil, xerrors.Errorf("Unable to create schema registry client: %w", err)
		}
		client.SetCredentials(in.Username, in.Password)
	}

	return debeziumengine.NewDebeziumImpl(logger, client, uint64(runtime.NumCPU()*4)), nil
}

func init() {
	parsers.Register(
		NewParserDebezium,
		[]parsers.AbstractParserConfig{new(ParserConfigDebeziumCommon), new(ParserConfigDebeziumLb)},
	)
}
