package bitbucketserver

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	bbv1 "github.com/gfleury/go-bitbucket-v1"
	"github.com/mitchellh/mapstructure"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/acl"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/params/info"
)

type activitiesTypes struct{ Values []*bbv1.Activity }

func (v *Provider) IsAllowed(ctx context.Context, event *info.Event) (bool, error) {
	allowed, err := v.checkMemberShip(event)
	if err != nil {
		return false, err
	}
	if allowed {
		return true, nil
	}

	// Check then from comment if there is a approved user that has done a /ok-to-test
	return v.checkOkToTestCommentFromApprovedMember(event)
}

func (v *Provider) isAllowedFromOwnerFile(event *info.Event) (bool, error) {
	ownerContent, err := v.GetFileInsideRepo(context.TODO(), event, "OWNERS", event.DefaultBranch)
	if err != nil {
		return false, err
	}

	return acl.UserInOwnerFile(ownerContent, event.AccountID)
}

func (v *Provider) checkOkToTestCommentFromApprovedMember(event *info.Event) (bool, error) {
	allPages, err := paginate(func(nextPage int) (*bbv1.APIResponse, error) {
		localVarOptionals := map[string]interface{}{
			"fromType": "COMMENT",
		}
		if nextPage > 0 {
			localVarOptionals["start"] = int(nextPage)
		}
		return v.Client.DefaultApi.GetActivities(v.projectKey, event.Repository, int64(v.pullRequestNumber), localVarOptionals)
	})
	if err != nil {
		return false, err
	}

	for _, comment := range allPages {
		activities := &activitiesTypes{}
		err := json.Unmarshal(comment.([]byte), activities)
		if err != nil {
			return false, err
		}
		for _, activity := range activities.Values {
			if acl.MatchRegexp(acl.OKToTestCommentRegexp, activity.Comment.Text) {
				commenterEvent := &info.Event{
					Sender:        activity.Comment.Author.Slug,
					AccountID:     fmt.Sprintf("%d", activity.Comment.Author.ID),
					Event:         event.Event,
					BaseBranch:    event.BaseBranch,
					HeadBranch:    event.HeadBranch,
					Repository:    event.Repository,
					Organization:  v.projectKey,
					DefaultBranch: event.DefaultBranch,
				}
				allowed, err := v.checkMemberShip(commenterEvent)
				if err != nil {
					return false, err
				}
				if allowed {
					// TODO: show a log how come this has been allowed
					return true, nil
				}
			}
		}
	}
	return false, nil
}

func (v *Provider) checkMemberShipResults(results []interface{}, event *info.Event) (bool, error) {
	accountintid, err := strconv.Atoi(event.AccountID)
	if err != nil {
		return false, err
	}
	for _, row := range results {
		user := &bbv1.UserPermission{}
		err := mapstructure.Decode(row, user)
		if err != nil {
			return false, err
		}

		if user.User.ID == accountintid {
			return true, nil
		}
	}
	return false, nil
}

func (v *Provider) checkMemberShip(event *info.Event) (bool, error) {
	// Get permissions from project
	allValues, err := paginate(func(nextPage int) (*bbv1.APIResponse, error) {
		localVarOptionals := map[string]interface{}{}
		if nextPage > 0 {
			localVarOptionals["start"] = int(nextPage)
		}
		return v.Client.DefaultApi.GetUsersWithAnyPermission_23(v.projectKey, localVarOptionals)
	})
	if err != nil {
		return false, err
	}
	allowed, err := v.checkMemberShipResults(allValues, event)
	if err != nil {
		return false, err
	}
	if allowed {
		return true, nil
	}

	// Get permissions from repo
	allValues, err = paginate(func(nextPage int) (*bbv1.APIResponse, error) {
		localVarOptionals := map[string]interface{}{}
		if nextPage > 0 {
			localVarOptionals["start"] = int(nextPage)
		}
		return v.Client.DefaultApi.GetUsersWithAnyPermission_24(v.projectKey, event.Repository, localVarOptionals)
	})
	if err != nil {
		return false, err
	}

	allowed, err = v.checkMemberShipResults(allValues, event)
	if err != nil {
		return false, err
	}
	if allowed {
		return true, nil
	}

	// Check if sender (which in bitbucket-server mean the accountID) is inside the Owner file
	// in the 'main' branch Silently ignore error, which should be fine it
	// probably means the OWNERS file is not created. If we had another error
	// (ie: like API) we probably would have hit it already.
	allowed, err = v.isAllowedFromOwnerFile(event)
	if allowed {
		return true, err
	}

	return false, nil
}
