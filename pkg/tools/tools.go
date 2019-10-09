package tools

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type RepoName struct {
	Tenant      string
	ProjectName string
	Stage       string
	Locator     string
}

var RepoParser = regexp.MustCompile("^(?P<tenant>[a-zA-Z0-9]*)-(?P<project>[a-zA-Z0-9-]*)-docker-(?P<stage>stable|staging)-(?P<locator>intranet|extranet)")

func RepoToStruct(repoName string) (*RepoName, error) {

	lowerRepoName := strings.ToLower(repoName)
	keys := RepoParser.SubexpNames()

	if len(keys) < 5 {
		return nil, errors.New(fmt.Sprintf(`
			The repository name parser doesn't have the four mandatory keys: tenant, project, stage and locator,
			you have only this: %v")
			 `, keys))
	}

	tenant, project, stage, locator := RepoParser.ReplaceAllString(lowerRepoName, "${tenant}"),
		RepoParser.ReplaceAllString(lowerRepoName, "${project}"),
		RepoParser.ReplaceAllString(lowerRepoName, "${stage}"),
		RepoParser.ReplaceAllString(lowerRepoName, "${locator}")

	repoFields := &RepoName{
		Tenant:      tenant,
		ProjectName: project,
		Stage:       stage,
		Locator:     locator,
	}

	return repoFields, nil
}
