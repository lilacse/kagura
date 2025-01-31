package dataservices

import (
	"context"

	"github.com/lilacse/kagura/dataservices/songdata"
)

type Provider struct {
	songdatasvc *songdata.Service
}

func NewProvider(ctx context.Context) (*Provider, error) {
	songdatasvc, err := songdata.NewService(ctx)
	if err != nil {
		return nil, err
	}

	return &Provider{
		songdatasvc: songdatasvc,
	}, nil
}

func (p *Provider) SongData() *songdata.Service {
	return p.songdatasvc
}
