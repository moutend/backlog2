//go:generate stringer -type=cacheType
package main

type cacheType int

const (
	IssueCommentsCache cacheType = iota
	IssueTypesCache
	IssuesCache
	IssueCache
	MyselfCache
	PrioritiesCache
	ProjectsCache
	ProjectCache
	PullRequestsCache
	RepositoriesCache
	StatusesCache
	WikisCache
	WikiCache
)
