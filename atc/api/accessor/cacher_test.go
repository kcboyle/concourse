package accessor_test

import (
	"github.com/concourse/concourse/atc/api/accessor"
	"github.com/concourse/concourse/atc/db"
	"github.com/concourse/concourse/atc/db/dbfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cacher", func() {
	var (
		fakeTeamFactory *dbfakes.FakeTeamFactory
		teamFetcher     accessor.TeamFetcher
		teams           []db.Team
		err             error
		fetchedTeams    []db.Team

		fakeTeam *dbfakes.FakeTeam
	)

	BeforeEach(func() {
		fakeTeam = new(dbfakes.FakeTeam)
		teams = []db.Team{fakeTeam}

		fakeTeamFactory = new(dbfakes.FakeTeamFactory)
		fakeTeamFactory.GetTeamsReturns(teams, nil)
	})

	JustBeforeEach(func() {
		teamFetcher = accessor.NewCacher(fakeTeamFactory)
		fetchedTeams, err = teamFetcher.GetTeams()
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when there is no cache found", func() {
		It("fetch teams from DB once", func() {
			Expect(fakeTeamFactory.GetTeamsCallCount()).To(Equal(1))
			Expect(fetchedTeams).To(Equal(teams))
		})
	})

	Context("when there is cache found", func() {
		JustBeforeEach(func() {
			fetchedTeams, err = teamFetcher.GetTeams()
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not fetch teams from DB again but read it from cache", func() {
			Expect(fakeTeamFactory.GetTeamsCallCount()).To(Equal(1))
			Expect(fetchedTeams).To(Equal(teams))
		})
	})

	Context("when unset teams", func() {
		JustBeforeEach(func() {
			teamFetcher.(accessor.Cacher).DeleteTeamsCache()
		})

		It("fetch teams again from DB since cache is not found", func() {
			fetchedTeams, err = teamFetcher.GetTeams()
			Expect(fakeTeamFactory.GetTeamsCallCount()).To(Equal(2))
		})
	})
})
