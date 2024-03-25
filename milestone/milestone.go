package milestone

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/go-github/v60/github"
)

// Find Milestone.
func Find(ctx context.Context, client *github.Client, owner, repositoryName string, major, minor int64) (*github.Milestone, error) {
	// TODO(ldez): get all the milestones.
	opt := &github.MilestoneListOptions{
		State:     "all",
		Direction: "desc",
	}

	milestones, _, err := client.Issues.ListMilestones(ctx, owner, repositoryName, opt)
	if err != nil {
		return nil, fmt.Errorf("list milestones: %w", err)
	}

	expectedTitle := strconv.FormatInt(major, 10) + "." + strconv.FormatInt(minor, 10)

	for _, milestone := range milestones {
		if strings.Contains(milestone.GetTitle(), expectedTitle) {
			fmt.Println(milestone.GetTitle())
			return milestone, nil
		}
	}
	return nil, fmt.Errorf("milestone not found: %s", expectedTitle)
}
