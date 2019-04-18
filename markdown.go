package main

import (
	"fmt"
	"io/ioutil"
	"net/url"

	"github.com/ericaro/frontmatter"
	backlog "github.com/moutend/go-backlog"
)

type issueFrontmatterOption struct {
	Summary   string `fm:"summary"`
	Project   string `fm:"project"`
	Parent    string `fm:"parent"`
	Type      string `fm:"type"`
	Priority  string `fm:"priority"`
	Status    string `fm:"status"`
	Start     string `fm:"start"`
	Due       string `fm:"due"`
	Estimated string `fm:"estimated"`
	Actual    string `fm:"actual"`
	Content   string `fm:"content"`
}

func parseIssueMarkdown(issueKey, path string) (url.Values, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var fo issueFrontmatterOption

	if err := frontmatter.Unmarshal(data, &fo); err != nil {
		return nil, err
	}

	var (
		myself      backlog.User
		project     backlog.Project
		issue       backlog.Issue
		parentIssue backlog.Issue
		issueType   backlog.IssueType
		priority    backlog.Priority
		status      backlog.Status
	)

	if err := fetchMyself(); err != nil {
		return nil, err
	}

	myself, err = readMyself()
	if err != nil {
		return nil, err
	}

	if err := fetchProjectByProjectKey(fo.Project); err != nil {
		return nil, err
	}

	project, err = readProjectByProjectKey(fo.Project)
	if err != nil {
		return nil, err
	}

	if issueKey != "" {
		if err := fetchIssue(issueKey); err != nil {
			return nil, err
		}

		issue, err = readIssue(issueKey)
		if err != nil {
			return nil, err
		}
	}

	if fo.Parent != "" {
		if err := fetchIssue(fo.Parent); err != nil {
			return nil, err
		}

		parentIssue, err = readIssue(fo.Parent)
		if err != nil {
			return nil, err
		}
	}
	if err := fetchIssueTypes(project.Id); err != nil {
		return nil, err
	}

	issueTypes, err := readIssueTypes(project.Id)
	if err != nil {
		return nil, err
	}
	for _, issueType = range issueTypes {
		if fo.Type == issueType.Name {
			break
		}
	}

	if err := fetchPriorities(); err != nil {
		return nil, err
	}

	priorities, err := readPriorities()
	if err != nil {
		return nil, err
	}
	for _, priority = range priorities {
		if fo.Priority == priority.Name {
			break
		}
	}

	if err := fetchStatuses(); err != nil {
		return nil, err
	}

	statuses, err := readStatuses()
	if err != nil {
		return nil, err
	}
	for _, status = range statuses {
		if fo.Status == status.Name {
			break
		}
	}

	values := url.Values{}
	values.Add("summary", fo.Summary)
	values.Add("description", fo.Content)
	values.Add("estimatedHours", fmt.Sprint(fo.Estimated))
	values.Add("actualHours", fmt.Sprint(fo.Actual))

	if issue.Id != 0 {
		values.Add("assigneeId", fmt.Sprint(issue.Assignee.Id))
	}
	if issueKey == "" {
		values.Add("assigneeId", fmt.Sprint(myself.Id))
	}
	if parentIssue.Id != 0 {
		values.Add("parentIssueId", fmt.Sprint(parentIssue.Id))

	}
	if fo.Due != "" {
		values.Add("dueDate", fo.Due)
	}
	if fo.Start != "" {
		values.Add("startDate", fo.Start)
	}

	return values, nil
}
