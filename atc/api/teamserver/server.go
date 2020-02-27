package teamserver

import (
	"code.cloudfoundry.org/lager"
	"github.com/concourse/concourse/atc/api/accessor"
	"github.com/concourse/concourse/atc/db"
)

type Server struct {
	logger      lager.Logger
	teamFactory db.TeamFactory
	externalURL string
	cacher      accessor.Cacher
}

func NewServer(
	logger lager.Logger,
	teamFactory db.TeamFactory,
	externalURL string,
	cacher accessor.Cacher,
) *Server {
	return &Server{
		logger:      logger,
		teamFactory: teamFactory,
		externalURL: externalURL,
		cacher:      cacher,
	}
}
