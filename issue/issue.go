package issue

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/go-github/github"
	"github.com/ldez/jaelon/types"
)

// AddMilestone on pull requests
func AddMilestone(ctx context.Context, client *github.Client, repoID types.RepoID, criteria types.SearchCriteria, mile *github.Milestone, verbose, dryRun bool) error {

	query, err := makeQuery(ctx, client, repoID, criteria, verbose)
	if err != nil {
		return err
	}

	searchOptions := &github.SearchOptions{
		Sort:        "created",
		Order:       "asc",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		// find Pull Request between currentRef and previousRef on baseBranch
		issuesSearchResult, resp, err := client.Search.Issues(ctx, query, searchOptions)
		if err != nil {
			return err
		}

		for _, issu := range issuesSearchResult.Issues {
			err = applyMilestone(ctx, client, repoID, issu, mile, dryRun)
			if err != nil {
				return err
			}
		}

		if resp.NextPage == 0 {
			break
		}
		searchOptions.Page = resp.NextPage
	}

	return nil
}

func makeQuery(ctx context.Context, client *github.Client, repoID types.RepoID, criteria types.SearchCriteria, verbose bool) (string, error) {

	// Get previous ref date
	commitPreviousRef, _, err := client.Repositories.GetCommit(ctx, repoID.Owner, repoID.RepositoryName, criteria.PreviousRef)
	if err != nil {
		return "", err
	}

	datePreviousRef := commitPreviousRef.Commit.Committer.GetDate().Add(1 * time.Second).Format("2006-01-02T15:04:05Z")

	// Get current ref version date
	commitCurrentRef, _, err := client.Repositories.GetCommit(ctx, repoID.Owner, repoID.RepositoryName, criteria.CurrentRef)
	if err != nil {
		return "", err
	}

	dateCurrentRef := commitCurrentRef.Commit.Committer.GetDate().Format("2006-01-02T15:04:05Z")

	// Search PR
	query := fmt.Sprintf("type:pr is:merged repo:%s/%s base:%s merged:%s..%s",
		repoID.Owner, repoID.RepositoryName, criteria.BaseBranch, datePreviousRef, dateCurrentRef)
	if verbose {
		log.Println(query)
	}

	return query, nil
}

func applyMilestone(ctx context.Context, client *github.Client, repoID types.RepoID, issue github.Issue, mile *github.Milestone, dryRun bool) error {
	if issue.Milestone == nil {
		log.Printf("No Milestone: https://github.com/%s/%s/pull/%d", repoID.Owner, repoID.RepositoryName, issue.GetNumber())
		if !dryRun {
			ir := &github.IssueRequest{
				Milestone: mile.Number,
			}
			_, _, err := client.Issues.Edit(ctx, repoID.Owner, repoID.RepositoryName, *issue.Number, ir)
			if err != nil {
				return err
			}
		}
	} else if issue.Milestone.GetID() == mile.GetID() {
		// no op
	} else {
		log.Printf("Milestone divergence: %s instead of %s, https://github.com/%s/%s/pull/%d",
			issue.Milestone.GetTitle(), mile.GetTitle(), repoID.Owner, repoID.RepositoryName, issue.GetNumber())
	}
	return nil
}
