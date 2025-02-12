// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package ttlcontroller

import (
	"context"
	"time"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

var (
	log = ctrl.Log.WithName("ttlcontroller")
)

// Controller defines the database ttl controller
type Controller struct {
	experiment core.ExperimentStore
	event      core.EventStore
	schedule   core.ScheduleStore
	workflow   core.WorkflowStore
	ttlconfig  *TTLconfig
}

// TTLconfig defines the ttl
type TTLconfig struct {
	// databaseTTLResyncPeriod defines the time interval to cleanup data in the database
	DatabaseTTLResyncPeriod time.Duration
	// EventTTL defines the ttl of events
	EventTTL time.Duration
	// ArchiveTTL defines the ttl of archives
	ArchiveTTL time.Duration
	// ScheduleTTL defines the ttl of schedule
	ScheduleTTL time.Duration
	// WorkflowTTL defines the ttl of workflow
	WorkflowTTL time.Duration
}

// NewController returns a new database ttl controller
func NewController(
	experiment core.ExperimentStore,
	event core.EventStore,
	schedule core.ScheduleStore,
	workflow core.WorkflowStore,
	ttlc *TTLconfig,
) *Controller {
	return &Controller{
		experiment: experiment,
		event:      event,
		schedule:   schedule,
		workflow:   workflow,
		ttlconfig:  ttlc,
	}
}

// Register periodically calls function runWorker to delete the data.
func Register(ctx context.Context, c *Controller) {
	defer utilruntime.HandleCrash()

	log.Info("Starting database TTL controller")

	go wait.Until(c.runWorker, c.ttlconfig.DatabaseTTLResyncPeriod, ctx.Done())
}

// runWorker is a long-running function that will be called in order to delete the events, archives, schedule, and workflow.
func (c *Controller) runWorker() {
	log.Info("Deleting expired data from the database")

	ctx := context.Background()

	_ = c.event.DeleteByDuration(ctx, c.ttlconfig.EventTTL)
	c.experiment.DeleteByFinishTime(ctx, c.ttlconfig.ArchiveTTL)
	c.schedule.DeleteByFinishTime(ctx, c.ttlconfig.ScheduleTTL)
	c.workflow.DeleteByEndTime(ctx, c.ttlconfig.WorkflowTTL)
}
