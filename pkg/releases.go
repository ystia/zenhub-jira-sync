package pkg

import (
	"log"

	"github.com/ystia/zenhub-jira-sync/pkg/clients/jira"
	"github.com/ystia/zenhub-jira-sync/pkg/clients/zenhub"
)

func (s *Sync) releases() ([]releasesTuple, error) {
	projectID, err := s.JiraClient.GetProjectID()
	if err != nil {
		return nil, err
	}
	log.Print("Listing ZenHub Releases Reports")
	zhReleases, err := s.ZenhubClient.GetReleasesReports()

	if err != nil {
		return nil, err
	}

	versions, err := s.JiraClient.GetProjectVersions()
	if err != nil {
		return nil, err
	}

	relTuples := make([]releasesTuple, 0)
	for _, release := range zhReleases {
		rExists := false
		if !s.ReleaseNameRE.MatchString(release.Title) {
			log.Printf("Ignoring ZenHub release %q that does not match the release name pattern", release.Title)
			continue
		}
		expectedVersionName := s.ReleaseNameRE.ReplaceAllString(release.Title, s.VersionNameRename)
		for _, version := range versions {
			if expectedVersionName == version.Name {
				rExists = true
				if diffReleaseAndVersion(release, version) {
					version, err = s.JiraClient.UpdateVersion(version)
					if err != nil {
						return nil, err
					}
				}
				relTuples = append(relTuples, releasesTuple{zhRelease: release, jiraVersion: version})
			}
		}
		if !rExists && release.State == "open" {
			version, err := s.JiraClient.CreateVersion(expectedVersionName, release.Description, projectID, release.State == "closed", false, release.StartDate, release.DesiredEndDate, release.ClosedAt)
			if err != nil {
				return nil, err
			}
			relTuples = append(relTuples, releasesTuple{zhRelease: release, jiraVersion: version})
		}
	}

	return relTuples, nil
}

func diffReleaseAndVersion(release *zenhub.ReleaseReport, version *jira.Version) bool {
	updatedVersion := false
	if release.State == "open" && version.Released {
		updatedVersion = true
		version.Released = false
	} else if release.State == "closed" && !version.Released {
		updatedVersion = true
		version.Released = true
	}
	if release.Description != version.Description {
		updatedVersion = true
		version.Description = release.Description
	}
	if release.StartDate != nil {
		rsd := release.StartDate.Format("2006-01-02")
		if version.StartDate != rsd {
			updatedVersion = true
			version.StartDate = rsd
		}
	}
	if release.DesiredEndDate != nil {
		ded := release.DesiredEndDate.Format("2/Jan/2006")
		if version.UserReleaseDate != ded {
			updatedVersion = true
			version.UserReleaseDate = ded
		}
	}
	if release.ClosedAt != nil {
		rca := release.ClosedAt.Format("2006-01-02")
		if version.ReleaseDate != rca {
			updatedVersion = true
			version.ReleaseDate = rca
		}
	}

	return updatedVersion
}
