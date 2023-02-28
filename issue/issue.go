package issue

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/go-github/v50/github"
	"github.com/ldez/jaelon/types"
)

// AddMilestone on pull requests.
func AddMilestone(ctx context.Context, client *github.Client, repoID types.RepoID, criteria types.SearchCriteria, mile *github.Milestone, verbose, dryRun bool) error {
	query, err := makeQuery(ctx, client, repoID, criteria, verbose)
	if err != nil {
		return fmt.Errorf("make query: %w", err)
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
				return fmt.Errorf("apply milestone: %w", err)
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
	if verbose {
		log.Printf("get previous commit [%s/%s@%s]", repoID.Owner, repoID.RepositoryName, criteria.PreviousRef)
	}
	commitPreviousRef, _, err := client.Repositories.GetCommit(ctx, repoID.Owner, repoID.RepositoryName, criteria.PreviousRef, nil)
	if err != nil {
		return "", fmt.Errorf("get previous commit [%s/%s@%s]: %w", repoID.Owner, repoID.RepositoryName, criteria.PreviousRef, err)
	}

	datePreviousRef := commitPreviousRef.Commit.Committer.GetDate().Add(1 * time.Second).Format("2006-01-02T15:04:05Z")

	// Get current ref version date
	if verbose {
		log.Printf("get previous commit [%s/%s@%s]", repoID.Owner, repoID.RepositoryName, criteria.CurrentRef)
	}
	commitCurrentRef, _, err := client.Repositories.GetCommit(ctx, repoID.Owner, repoID.RepositoryName, criteria.CurrentRef, nil)
	if err != nil {
		return "", fmt.Errorf("get current commit [%s/%s@%s]: %w", repoID.Owner, repoID.RepositoryName, criteria.CurrentRef, err)
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

func applyMilestone(ctx context.Context, client *github.Client, repoID types.RepoID, issue *github.Issue, mile *github.Milestone, dryRun bool) error {
	switch {
	case issue.Milestone == nil:
		log.Printf("%s No Milestone: https://github.com/%s/%s/pull/%d", issue.GetClosedAt(), repoID.Owner, repoID.RepositoryName, issue.GetNumber())
		if !dryRun {
			ir := &github.IssueRequest{
				Milestone: mile.Number,
			}
			_, _, err := client.Issues.Edit(ctx, repoID.Owner, repoID.RepositoryName, *issue.Number, ir)
			if err != nil {
				return fmt.Errorf("issue edit: %w", err)
			}
		}
	case issue.Milestone.GetID() == mile.GetID():
		// no op
	default:
		log.Printf("Milestone divergence: %s instead of %s, https://github.com/%s/%s/pull/%d",
			issue.Milestone.GetTitle(), mile.GetTitle(), repoID.Owner, repoID.RepositoryName, issue.GetNumber())
	}

	return nil
}
