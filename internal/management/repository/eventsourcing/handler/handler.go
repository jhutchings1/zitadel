package handler

import (
	"time"

	"github.com/caos/zitadel/internal/config/types"
	"github.com/caos/zitadel/internal/eventstore"
	"github.com/caos/zitadel/internal/eventstore/query"
	"github.com/caos/zitadel/internal/management/repository/eventsourcing/view"
	org_event "github.com/caos/zitadel/internal/org/repository/eventsourcing"
	proj_event "github.com/caos/zitadel/internal/project/repository/eventsourcing"
	usr_event "github.com/caos/zitadel/internal/user/repository/eventsourcing"
)

type Configs map[string]*Config

type Config struct {
	MinimumCycleDuration types.Duration
}

type handler struct {
	view                *view.View
	bulkLimit           uint64
	cycleDuration       time.Duration
	errorCountUntilSkip uint64
}

type EventstoreRepos struct {
	ProjectEvents *proj_event.ProjectEventstore
	UserEvents    *usr_event.UserEventstore
	OrgEvents     *org_event.OrgEventstore
}

func Register(configs Configs, bulkLimit, errorCount uint64, view *view.View, eventstore eventstore.Eventstore, repos EventstoreRepos) []query.Handler {
	return []query.Handler{
		&Project{handler: handler{view, bulkLimit, configs.cycleDuration("Project"), errorCount}, eventstore: eventstore},
		&ProjectGrant{handler: handler{view, bulkLimit, configs.cycleDuration("ProjectGrant"), errorCount}, eventstore: eventstore, projectEvents: repos.ProjectEvents, orgEvents: repos.OrgEvents},
		&ProjectRole{handler: handler{view, bulkLimit, configs.cycleDuration("ProjectRole"), errorCount}, projectEvents: repos.ProjectEvents},
		&ProjectMember{handler: handler{view, bulkLimit, configs.cycleDuration("ProjectMember"), errorCount}, userEvents: repos.UserEvents},
		&ProjectGrantMember{handler: handler{view, bulkLimit, configs.cycleDuration("ProjectGrantMember"), errorCount}, userEvents: repos.UserEvents},
		&Application{handler: handler{view, bulkLimit, configs.cycleDuration("Application"), errorCount}, projectEvents: repos.ProjectEvents},
		&User{handler: handler{view, bulkLimit, configs.cycleDuration("User"), errorCount}, eventstore: eventstore, orgEvents: repos.OrgEvents},
		&UserGrant{handler: handler{view, bulkLimit, configs.cycleDuration("UserGrant"), errorCount}, projectEvents: repos.ProjectEvents, userEvents: repos.UserEvents, orgEvents: repos.OrgEvents},
		&Org{handler: handler{view, bulkLimit, configs.cycleDuration("Org"), errorCount}},
		&OrgMember{handler: handler{view, bulkLimit, configs.cycleDuration("OrgMember"), errorCount}, userEvents: repos.UserEvents},
		&OrgDomain{handler: handler{view, bulkLimit, configs.cycleDuration("OrgDomain"), errorCount}},
	}
}

func (configs Configs) cycleDuration(viewModel string) time.Duration {
	c, ok := configs[viewModel]
	if !ok {
		return 1 * time.Second
	}
	return c.MinimumCycleDuration.Duration
}

func (h *handler) MinimumCycleDuration() time.Duration {
	return h.cycleDuration
}

func (h *handler) QueryLimit() uint64 {
	return h.bulkLimit
}
