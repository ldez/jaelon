package types

// Configuration Jaelon configuration.
type Configuration struct {
	CurrentVersionTemplate  string
	PreviousVersionTemplate string
	ReleaseBranchTemplate   string
	BaseBranch              string
	Major                   int64
	Minor                   int64
	Current                 bool
	Owner                   string
	RepositoryName          string
	GitHubToken             string
	Debug                   bool
	DryRun                  bool
}

// RepoID GitHub repo ID.
type RepoID struct {
	Owner          string
	RepositoryName string
}

// SearchCriteria search criterion.
type SearchCriteria struct {
	CurrentRef  string
	PreviousRef string
	BaseBranch  string
}
