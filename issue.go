package main

import (
	"github.com/google/go-github/github"
	"fmt"
)

// ListIssuesByRepo lists issues within given repo, filters by labels if provided
func (gc *GithubClient) ListIssuesByRepo(org, repo string, labels []string) ([]*github.Issue, error) {
	issueListOptions := github.IssueListByRepoOptions{
		State: string(IssueAllState),
	}
	if len(labels) > 0 {
		issueListOptions.Labels = labels
	}

	var res []*github.Issue
	options := &github.ListOptions{}
	genericList, err := gc.depaginate(
		fmt.Sprintf("listing issues with label '%v'", labels),
		maxRetryCount,
		options,
		func() ([]interface{}, *github.Response, error) {
			page, resp, err := gc.Client.Issues.ListByRepo(ctx, org, repo, &issueListOptions)
			var interfaceList []interface{}
			if nil == err {
				for _, issue := range page {
					interfaceList = append(interfaceList, issue)
				}
			}
			return interfaceList, resp, err
		},
	)
	for _, issue := range genericList {
		res = append(res, issue.(*github.Issue))
	}
	return res, err
}

// CreateIssue creates issue
func (gc *GithubClient) CreateIssue(org, repo, title, body string) (*github.Issue, error) {
	issue := &github.IssueRequest{
		Title: &title,
		Body:  &body,
	}

	var res *github.Issue
	_, err := gc.retry(
		fmt.Sprintf("creating issue '%s %s' '%s'", org, repo, title),
		maxRetryCount,
		func() (*github.Response, error) {
			var resp *github.Response
			var err error
			res, resp, err = gc.Client.Issues.Create(ctx, org, repo, issue)
			return resp, err
		},
	)
	return res, err
}

// CloseIssue closes issue
func (gc *GithubClient) CloseIssue(org, repo string, issueNumber int) error {
	return gc.updateIssueState(org, repo, IssueCloseState, issueNumber)
}

// ReopenIssue reopen issue
func (gc *GithubClient) ReopenIssue(org, repo string, issueNumber int) error {
	return gc.updateIssueState(org, repo, IssueOpenState, issueNumber)
}

// AddLabelsToIssue adds label on issue
func (gc *GithubClient) AddLabelsToIssue(org, repo string, issueNumber int, labels []string) error {
	_, err := gc.retry(
		fmt.Sprintf("add labels '%v' to '%s %s %d'", labels, org, repo, issueNumber),
		maxRetryCount,
		func() (*github.Response, error) {
			_, resp, err := gc.Client.Issues.AddLabelsToIssue(ctx, org, repo, issueNumber, labels)
			return resp, err
		},
	)
	return err
}

// RemoveLabelForIssue removes given label for issue
func (gc *GithubClient) RemoveLabelForIssue(org, repo string, issueNumber int, label string) error {
	_, err := gc.retry(
		fmt.Sprintf("remove label '%s' from '%s %s %d'", label, org, repo, issueNumber),
		maxRetryCount,
		func() (*github.Response, error) {
			return gc.Client.Issues.RemoveLabelForIssue(ctx, org, repo, issueNumber, label)
		},
	)
	return err
}

func (gc *GithubClient) updateIssueState(org, repo string, state IssueStateEnum, issueNumber int) error {
	stateString := string(state)
	issueRequest := &github.IssueRequest{
		State: &stateString,
	}
	_, err := gc.retry(
		fmt.Sprintf("applying '%s' action on issue '%s %s %d'", stateString, org, repo, issueNumber),
		maxRetryCount,
		func() (*github.Response, error) {
			_, resp, err := gc.Client.Issues.Edit(ctx, org, repo, issueNumber, issueRequest)
			return resp, err
		},
	)
	return err
}
