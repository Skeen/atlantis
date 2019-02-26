// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

package events_test

import (
	"errors"
	"strings"
	"testing"

	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs/mocks"
	. "github.com/runatlantis/atlantis/testing"
)

var repoModel = models.Repo{}
var pullModel = models.PullRequest{}
var status = models.SuccessCommitStatus

func TestUpdate(t *testing.T) {
	RegisterMockTestingT(t)
	client := mocks.NewMockClient()
	s := events.DefaultCommitStatusUpdater{Client: client}
	err := s.Update(repoModel, pullModel, status, models.PlanCommand)
	Ok(t, err)
	client.VerifyWasCalledOnce().UpdateStatus(repoModel, pullModel, status, "Atlantis", "Plan Success")
}

func TestUpdateProjectResult_Error(t *testing.T) {
	RegisterMockTestingT(t)
	ctx := &events.CommandContext{
		BaseRepo: repoModel,
		Pull:     pullModel,
	}
	client := mocks.NewMockClient()
	s := events.DefaultCommitStatusUpdater{Client: client}
	err := s.UpdateProjectResult(ctx, models.PlanCommand, events.CommandResult{Error: errors.New("err")})
	Ok(t, err)
	client.VerifyWasCalledOnce().UpdateStatus(repoModel, pullModel, models.FailedCommitStatus, "Atlantis", "Plan Failed")
}

func TestUpdateProjectResult_Failure(t *testing.T) {
	RegisterMockTestingT(t)
	ctx := &events.CommandContext{
		BaseRepo: repoModel,
		Pull:     pullModel,
	}
	client := mocks.NewMockClient()
	s := events.DefaultCommitStatusUpdater{Client: client}
	err := s.UpdateProjectResult(ctx, models.PlanCommand, events.CommandResult{Failure: "failure"})
	Ok(t, err)
	client.VerifyWasCalledOnce().UpdateStatus(repoModel, pullModel, models.FailedCommitStatus, "Atlantis", "Plan Failed")
}

func TestUpdateProjectResult(t *testing.T) {
	RegisterMockTestingT(t)

	ctx := &events.CommandContext{
		BaseRepo: repoModel,
		Pull:     pullModel,
	}

	cases := []struct {
		Statuses []string
		Expected models.CommitStatus
	}{
		{
			[]string{"success", "failure", "error"},
			models.FailedCommitStatus,
		},
		{
			[]string{"failure", "error", "success"},
			models.FailedCommitStatus,
		},
		{
			[]string{"success", "failure"},
			models.FailedCommitStatus,
		},
		{
			[]string{"success", "error"},
			models.FailedCommitStatus,
		},
		{
			[]string{"failure", "error"},
			models.FailedCommitStatus,
		},
		{
			[]string{"success"},
			models.SuccessCommitStatus,
		},
		{
			[]string{"success", "success"},
			models.SuccessCommitStatus,
		},
	}

	for _, c := range cases {
		t.Run(strings.Join(c.Statuses, "-"), func(t *testing.T) {
			var results []models.ProjectResult
			for _, statusStr := range c.Statuses {
				var result models.ProjectResult
				switch statusStr {
				case "failure":
					result = models.ProjectResult{
						Failure: "failure",
					}
				case "error":
					result = models.ProjectResult{
						Error: errors.New("err"),
					}
				default:
					result = models.ProjectResult{}
				}
				results = append(results, result)
			}
			resp := events.CommandResult{ProjectResults: results}

			client := mocks.NewMockClient()
			s := events.DefaultCommitStatusUpdater{Client: client}
			err := s.UpdateProjectResult(ctx, models.PlanCommand, resp)
			Ok(t, err)
			client.VerifyWasCalledOnce().UpdateStatus(repoModel, pullModel, c.Expected, "Atlantis", "Plan "+strings.Title(c.Expected.String()))
		})
	}
}
