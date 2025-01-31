package dataservices

import (
	"context"

	"github.com/lilacse/kagura/dataservices/songdata"
)

type Provider struct {
	songdatasvc *songdata.SongDataService
}

func NewProvider(ctx context.Context) (*Provider, error) {
	songdatasvc, err := songdata.NewSongDataService(ctx)
	if err != nil {
		return nil, err
	}

	return &Provider{
		songdatasvc: songdatasvc,
	}, nil
}

func (p *Provider) SongData() *songdata.SongDataService {
	return p.songdatasvc
}
