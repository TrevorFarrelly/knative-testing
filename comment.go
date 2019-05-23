package main

import (
	"github.com/google/go-github/github"
	"fmt"
)

// ListComments gets all comments from issue
func (gc *GithubClient) ListComments(org, repo string, issueNumber int) ([]*github.IssueComment, error) {
	var res []*github.IssueComment
	options := &github.ListOptions{}
	genericList, err := gc.depaginate(
		fmt.Sprintf("listing comment for issue '%s %s %d'", org, repo, issueNumber),
		maxRetryCount,
		options,
		func() ([]interface{}, *github.Response, error) {
			page, resp, err := gc.Client.Issues.ListComments(ctx, org, repo, issueNumber, nil)
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
		res = append(res, issue.(*github.IssueComment))
	}
	return res, err
}

// GetComment gets comment by comment ID
func (gc *GithubClient) GetComment(org, repo string, commentID int64) (*github.IssueComment, error) {
	var res *github.IssueComment
	_, err := gc.retry(
		fmt.Sprintf("getting comment '%s %s %d'", org, repo, commentID),
		maxRetryCount,
		func() (*github.Response, error) {
			var resp *github.Response
			var err error
			res, resp, err = gc.Client.Issues.GetComment(ctx, org, repo, commentID)
			return resp, err
		},
	)
	return res, err
}

// CreateComment adds comment to issue
func (gc *GithubClient) CreateComment(org, repo string, issueNumber int, commentBody string) (*github.IssueComment, error) {
	var res *github.IssueComment
	comment := &github.IssueComment{
		Body: &commentBody,
	}
	_, err := gc.retry(
		fmt.Sprintf("commenting issue '%s %s %d'", org, repo, issueNumber),
		maxRetryCount,
		func() (*github.Response, error) {
			var resp *github.Response
			var err error
			res, resp, err = gc.Client.Issues.CreateComment(ctx, org, repo, issueNumber, comment)
			return resp, err
		},
	)
	return res, err
}

// EditComment edits comment by replacing with provided comment
func (gc *GithubClient) EditComment(org, repo string, commentID int64, commentBody string) error {
	comment := &github.IssueComment{
		Body: &commentBody,
	}
	_, err := gc.retry(
		fmt.Sprintf("editing comment '%s %s %d'", org, repo, commentID),
		maxRetryCount,
		func() (*github.Response, error) {
			_, resp, err := gc.Client.Issues.EditComment(ctx, org, repo, commentID, comment)
			return resp, err
		},
	)
	return err
}
