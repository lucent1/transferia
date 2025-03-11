package provider

import (
	"context"

	"github.com/transferria/transferria/library/go/core/metrics"
	"github.com/transferria/transferria/library/go/core/xerrors"
	"github.com/transferria/transferria/pkg/abstract"
	cpclient "github.com/transferria/transferria/pkg/abstract/coordinator"
	"github.com/transferria/transferria/pkg/abstract/model"
	"github.com/transferria/transferria/pkg/middlewares"
	"github.com/transferria/transferria/pkg/providers"
	"github.com/transferria/transferria/pkg/providers/s3"
	_ "github.com/transferria/transferria/pkg/providers/s3/fallback"
	s3_sink "github.com/transferria/transferria/pkg/providers/s3/sink"
	"github.com/transferria/transferria/pkg/providers/s3/source"
	"github.com/transferria/transferria/pkg/providers/s3/storage"
	"go.ytsaurus.tech/library/go/core/log"
)

func init() {
	providers.Register(s3.ProviderType, New)
}

// To verify providers contract implementation
var (
	_ providers.Sinker      = (*Provider)(nil)
	_ providers.Snapshot    = (*Provider)(nil)
	_ providers.Activator   = (*Provider)(nil)
	_ providers.Replication = (*Provider)(nil)
)

type Provider struct {
	logger   log.Logger
	registry metrics.Registry
	cp       cpclient.Coordinator
	transfer *model.Transfer
}

func (p *Provider) Activate(ctx context.Context, task *model.TransferOperation, tables abstract.TableMap, callbacks providers.ActivateCallbacks) error {
	if !p.transfer.IncrementOnly() {
		if err := callbacks.Cleanup(tables); err != nil {
			return xerrors.Errorf("Sinker cleanup failed: %w", err)
		}
		if err := callbacks.CheckIncludes(tables); err != nil {
			return xerrors.Errorf("Failed in accordance with configuration: %w", err)
		}
		if err := callbacks.Upload(tables); err != nil {
			return xerrors.Errorf("Snapshot loading failed: %w", err)
		}
	}
	return nil
}

func (p *Provider) Storage() (abstract.Storage, error) {
	src, ok := p.transfer.Src.(*s3.S3Source)
	if !ok {
		return nil, xerrors.Errorf("unexpected source type: %T", p.transfer.Src)
	}
	return storage.New(src, p.logger, p.registry)
}

func (p *Provider) Type() abstract.ProviderType {
	return s3.ProviderType
}

func (p *Provider) Source() (abstract.Source, error) {
	src, ok := p.transfer.Src.(*s3.S3Source)
	if !ok {
		return nil, xerrors.Errorf("unexpected source type: %T", p.transfer.Src)
	}
	return source.NewSource(src, p.transfer.ID, p.logger, p.registry, p.cp)
}

func (p *Provider) Sink(middlewares.Config) (abstract.Sinker, error) {
	dst, ok := p.transfer.Dst.(*s3.S3Destination)
	if !ok {
		return nil, xerrors.Errorf("unexpected target type: %T", p.transfer.Dst)
	}
	return s3_sink.NewSinker(p.logger, dst, p.registry, p.cp, p.transfer.ID)
}

func New(lgr log.Logger, registry metrics.Registry, cp cpclient.Coordinator, transfer *model.Transfer) providers.Provider {
	return &Provider{
		logger:   lgr,
		registry: registry,
		cp:       cp,
		transfer: transfer,
	}
}
