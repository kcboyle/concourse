package accessor

import (
	"time"

	"github.com/concourse/concourse/atc/db"
	"github.com/patrickmn/go-cache"
)

type Cacher interface {
	DeleteTeamsCache()
}

type cacher struct {
	cache       *cache.Cache
	teamFactory db.TeamFactory
}

func NewCacher(teamFactory db.TeamFactory) *cacher {
	return &cacher{
		cache:       cache.New(time.Minute, time.Minute),
		teamFactory: teamFactory,
	}
}

func (c *cacher) GetTeams() ([]db.Team, error) {
	if teams, found := c.cache.Get("teams"); found {
		return teams.([]db.Team), nil
	}

	teams, err := c.teamFactory.GetTeams()
	if err != nil {
		return nil, err
	}

	c.cache.Set("teams", teams, cache.DefaultExpiration)

	return teams, nil
}

func (c *cacher) DeleteTeamsCache() {
	c.cache.Delete("teams")
}
