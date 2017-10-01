package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/containous/flaeg"
	"github.com/google/go-github/github"
	"github.com/ldez/jaelon/issue"
	"github.com/ldez/jaelon/milestone"
	"github.com/ldez/jaelon/types"
	"golang.org/x/oauth2"
)

func main() {
	config := &types.Configuration{
		CurrentVersionTemplate:  "v%v.%v.0-rc1",
		PreviousVersionTemplate: "v%v.%v.0-rc1",
		ReleaseBranchTemplate:   "v%v.%v",
		BaseBranch:              "master",
		Major:                   1,
		Minor:                   0,
		DryRun:                  true,
	}

	defaultConfig := &types.Configuration{}

	rootCmd := &flaeg.Command{
		Name: "jaelon",
		Description: `Jaelon is a GitHub Milestone checker and fixer.
Check if Pull Requests have a Milestone.`,
		DefaultPointersConfig: defaultConfig,
		Config:                config,
		Run:                   runCmd(config),
	}

	flag := flaeg.New(rootCmd, os.Args[1:])
	flag.Run()
}

func runCmd(config *types.Configuration) func() error {
	return func() error {
		if config.Debug {
			log.Printf("Run Jaelon command with config : %+v\n", config)
		}

		if config.DryRun {
			log.Print("IMPORTANT: you are using the dry-run mode. Use `--dry-run=false` to disable this mode.")
		}

		err := required(config.Owner, "owner")
		if err != nil {
			log.Fatal(err)
		}
		err = required(config.RepositoryName, "repo-name")
		if err != nil {
			log.Fatal(err)
		}

		ctx := context.Background()
		client := newGitHubClient(ctx, config.GitHubToken)

		err = browse(ctx, client, config)
		if err != nil {
			log.Fatal(err)
		}
		return nil
	}
}

func browse(ctx context.Context, client *github.Client, config *types.Configuration) error {

	repoID := types.RepoID{
		Owner:          config.Owner,
		RepositoryName: config.RepositoryName,
	}

	mile, err := milestone.Find(ctx, client, repoID.Owner, repoID.RepositoryName, config.Major, config.Minor)
	if err != nil {
		return err
	}

	criterion := []types.SearchCriteria{
		makeSearchCriteriaMaster(config),
	}

	if !config.Current {
		criteria := makeSearchCriteriaVersionBranch(config)
		criterion = append(criterion, criteria)
	}

	for _, criteria := range criterion {
		err = issue.AddMilestone(ctx, client, repoID, criteria, mile, config.Debug, config.DryRun)
		if err != nil {
			return err
		}
	}

	return nil
}

func makeSearchCriteriaMaster(config *types.Configuration) types.SearchCriteria {
	var currentRef string
	if config.Current {
		currentRef = config.BaseBranch
	} else {
		currentRef = fmt.Sprintf(config.CurrentVersionTemplate, config.Major, config.Minor)
	}

	return types.SearchCriteria{
		BaseBranch:  config.BaseBranch,
		CurrentRef:  currentRef,
		PreviousRef: getPreviousRef(config),
	}
}

func makeSearchCriteriaVersionBranch(config *types.Configuration) types.SearchCriteria {
	baseBranch := fmt.Sprintf(config.ReleaseBranchTemplate, config.Major, config.Minor)

	return types.SearchCriteria{
		BaseBranch:  baseBranch,
		CurrentRef:  baseBranch,
		PreviousRef: getPreviousRef(config),
	}
}

func getPreviousRef(config *types.Configuration) string {
	return fmt.Sprintf(config.PreviousVersionTemplate, config.Major, config.Minor-1)
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

func required(field string, fieldName string) error {
	if len(field) == 0 {
		log.Fatalf("%s is mandatory.", fieldName)
	}
	return nil
}
