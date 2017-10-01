package types

// Configuration Jaelon configuration
type Configuration struct {
	CurrentVersionTemplate  string
	PreviousVersionTemplate string
	ReleaseBranchTemplate   string
	BaseBranch              string
	Major                   int64  `short:"a" description:"Major version part of the Milestone."`
	Minor                   int64  `short:"i" description:"Minor version part of the Milestone."`
	Current                 bool   `short:"c" description:"Follow the head of master."`
	Owner                   string `short:"o" description:"Repository owner."`
	RepositoryName          string `long:"repo-name" short:"r" description:"Repository name."`
	GitHubToken             string `long:"token" short:"t" description:"GitHub Token."`
	Debug                   bool   `long:"debug" description:"Debug mode."`
	DryRun                  bool   `long:"dry-run" description:"Dry run mode."`
}

// RepoID GitHub repo ID
type RepoID struct {
	Owner          string
	RepositoryName string
}

// SearchCriteria search criterion
type SearchCriteria struct {
	CurrentRef  string
	PreviousRef string
	BaseBranch  string
}
