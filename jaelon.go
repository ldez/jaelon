package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type Configuration struct {
	Major          int64  `short:"a" description:"TODO"`
	Minor          int64  `short:"i" description:"TODO"`
	Current        bool   `short:"c" description:"TODO"`
	Owner          string `short:"o" description:"Repository owner."`
	RepositoryName string `long:"repo-name" short:"r" description:"Repository name."`
	GitHubToken    string `long:"token" short:"t" description:"GitHub Token."`
	Debug          bool   `long:"debug" description:"Debug mode."`
}

func main() {
	config := &Configuration{
		Major:          1,
		Minor:          3,
		Current:        false,
		Owner:          "containous",
		RepositoryName: "traefik",
		GitHubToken:    "",
		Debug:          true,
	}
	browse(config)
}

func browse(config *Configuration) {
	ctx := context.Background()

	client := newGitHubClient(ctx, config.GitHubToken)

	milestone, err := findMilestone(ctx, client, config)
	check(err)

	// find on master
	baseBranch := "master"
	previousRef := fmt.Sprintf("v%v.%v.0-rc1", config.Major, config.Minor-1)

	var currentRef string
	if config.Current {
		currentRef = baseBranch
	} else {
		currentRef = fmt.Sprintf("v%v.%v.0-rc1", config.Major, config.Minor)
	}

	prOnMaster := findIssues(ctx, client, config, currentRef, previousRef, baseBranch)
	checkMilestone(prOnMaster, milestone)

	// find on version branch
	if !config.Current {
		baseBranch = fmt.Sprintf("v%v.%v", config.Major, config.Minor)
		currentRef = baseBranch

		prOnBranch := findIssues(ctx, client, config, currentRef, previousRef, baseBranch)
		checkMilestone(prOnBranch, milestone)
	}
}
func findIssues(ctx context.Context, client *github.Client, config *Configuration, currentRef string, previousRef string, baseBranch string) []github.Issue {

	// Get previous ref date
	commitPreviousRef, _, err := client.Repositories.GetCommit(ctx, config.Owner, config.RepositoryName, previousRef)
	check(err)

	datePreviousRef := commitPreviousRef.Commit.Committer.Date.Add(1 * time.Second).Format("2006-01-02T15:04:05Z")

	// Get current ref version date
	commitCurrentRef, _, err := client.Repositories.GetCommit(ctx, config.Owner, config.RepositoryName, currentRef)
	check(err)

	dateCurrentRef := commitCurrentRef.Commit.Committer.Date.Format("2006-01-02T15:04:05Z")

	// Search PR
	query := fmt.Sprintf("type:pr is:merged repo:%s/%s base:%s merged:%s..%s",
		config.Owner, config.RepositoryName, baseBranch, datePreviousRef, dateCurrentRef)
	if config.Debug {
		log.Println(query)
	}

	searchOptions := &github.SearchOptions{
		Sort:        "created",
		Order:       "asc",
		ListOptions: github.ListOptions{PerPage: 20},
	}

	return searchAllIssues(ctx, client, query, searchOptions)
}

func checkMilestone(allSearchResult []github.Issue, milestone *github.Milestone) {
	for _, issue := range allSearchResult {
		if issue.Milestone == nil {
			log.Printf("No Milestone: #%v", *issue.Number)
			//ir := &github.IssueRequest{
			//	Milestone: milestone.ID,
			//}
			//_, _, err = client.Issues.Edit(ctx, config.Owner, config.RepositoryName, *issue.Number, ir)
			//check(err)
		} else if *issue.Milestone.ID == *milestone.ID {
			// no op
		} else {
			log.Printf("Milestone divergence: #%v. %s instead of %s", *issue.Number, *issue.Milestone.Title, *milestone.Title)
		}
	}
}

func findMilestone(ctx context.Context, client *github.Client, config *Configuration) (*github.Milestone, error) {
	opt := &github.MilestoneListOptions{
		State: "all",
	}

	milestones, _, err := client.Issues.ListMilestones(ctx, config.Owner, config.RepositoryName, opt)
	check(err)

	expectedTitle := strconv.FormatInt(config.Major, 10) + "." + strconv.FormatInt(config.Minor, 10)

	for _, milestone := range milestones {
		if strings.Contains(*milestone.Title, expectedTitle) {
			fmt.Println(*milestone.Title)
			return milestone, nil
		}
	}
	return nil, fmt.Errorf("Milestone not found: %s", expectedTitle)
}

func searchAllIssues(ctx context.Context, client *github.Client, query string, searchOptions *github.SearchOptions) []github.Issue {
	var allSearchResult []github.Issue
	for {
		issuesSearchResult, resp, err := client.Search.Issues(ctx, query, searchOptions)
		if err != nil {
			log.Fatal(err)
		}
		for _, issueResult := range issuesSearchResult.Issues {
			allSearchResult = append(allSearchResult, issueResult)
		}
		if resp.NextPage == 0 {
			break
		}
		searchOptions.Page = resp.NextPage
	}
	return allSearchResult
}

func newGitHubClient(ctx context.Context, token string) *github.Client {
	var client *github.Client
	if len(token) == 0 {
		client = github.NewClient(nil)
	} else {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	}
	return client
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
