package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ettle/strcase"
	"github.com/google/go-github/v50/github"
	"github.com/ldez/jaelon/issue"
	"github.com/ldez/jaelon/milestone"
	"github.com/ldez/jaelon/types"
	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2"
)

const (
	flagDebug       = "debug"
	flagCurrent     = "current"
	flagDryRun      = "dry-run"
	flagMajor       = "major"
	flagMinor       = "minor"
	flagOwner       = "owner"
	flagRepoName    = "repo-name"
	flagGitHubToken = "github-token"
)

func getFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:    "c",
			Usage:   "Follow the head of master.",
			EnvVars: []string{strcase.ToSNAKE(flagCurrent)},
			Aliases: []string{flagCurrent},
		},
		&cli.BoolFlag{
			Name:    flagDebug,
			Usage:   "Debug mode.",
			EnvVars: []string{strcase.ToSNAKE(flagDebug)},
		},
		&cli.BoolFlag{
			Name:    flagDryRun,
			Usage:   "Debug mode.",
			Value:   true,
			EnvVars: []string{strcase.ToSNAKE(flagDryRun)},
		},
		&cli.Int64Flag{
			Name:    "a",
			Usage:   "Major version part of the Milestone.",
			Value:   1,
			EnvVars: []string{strcase.ToSNAKE(flagMajor)},
			Aliases: []string{flagMajor},
		},
		&cli.Int64Flag{
			Name:    "i",
			Usage:   "Minor version part of the Milestone.",
			Value:   0,
			EnvVars: []string{strcase.ToSNAKE(flagMinor)},
			Aliases: []string{flagMinor},
		},
		&cli.StringFlag{
			Name:     "o",
			Usage:    "Repository owner.",
			EnvVars:  []string{strcase.ToSNAKE(flagOwner)},
			Aliases:  []string{flagOwner},
			Required: true,
		},
		&cli.StringFlag{
			Name:     "r",
			Usage:    "Repository name.",
			EnvVars:  []string{strcase.ToSNAKE(flagRepoName)},
			Aliases:  []string{flagRepoName},
			Required: true,
		},
		&cli.StringFlag{
			Name:    "t",
			Usage:   "GitHub Token.",
			EnvVars: []string{strcase.ToSNAKE(flagGitHubToken)},
			Aliases: []string{flagGitHubToken, "token"},
		},
	}
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

	err := run()
	if err != nil {
		log.Fatalf("Error while executing command: %v", err)
	}
}

func run() error {
	app := &cli.App{
		Name: "jaelon",
		Usage: `GitHub Milestone checker and fixer.
Check if Pull Requests have a Milestone.`,
		Flags: getFlags(),
		Action: func(c *cli.Context) error {
			config := &types.Configuration{
				CurrentVersionTemplate:  "v%v.%v.0",
				PreviousVersionTemplate: "v%v.%v.0",
				ReleaseBranchTemplate:   "v%v.%v",
				BaseBranch:              "master",
				Major:                   c.Int64(flagMajor),
				Minor:                   c.Int64(flagMinor),
				Current:                 c.Bool(flagCurrent),
				Owner:                   c.String(flagOwner),
				RepositoryName:          c.String(flagRepoName),
				GitHubToken:             c.String(flagGitHubToken),
				Debug:                   c.Bool(flagDebug),
				DryRun:                  c.Bool(flagDryRun),
			}

			return runCmd(config)
		},
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	err := app.RunContext(ctx, os.Args)
	if err != nil {
		return err
	}

	return nil
}

func runCmd(config *types.Configuration) error {
	if config.Debug {
		log.Printf("Run Jaelon command with config : %+v\n", config)
	}

	if config.DryRun {
		log.Print("IMPORTANT: you are using the dry-run mode. Use `--dry-run=false` to disable this mode.")
	}

	ctx := context.Background()
	client := newGitHubClient(ctx, config.GitHubToken)

	return browse(ctx, client, config)
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
	if token == "" {
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
