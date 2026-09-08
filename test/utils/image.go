package utils

import "strings"

// ImageRepository returns the repository portion of a container image
// reference, with any tag and/or digest stripped.
//
// It only looks for the tag separator (":") after the last "/", so a
// registry host with a port (e.g. "myregistry:5000/operator:1.2.3") isn't
// mistaken for a tag.
func ImageRepository(image string) string {
	repo := image
	if i := strings.LastIndex(repo, "/"); i >= 0 {
		host, rest := repo[:i], repo[i:]
		if j := strings.Index(rest, ":"); j >= 0 {
			rest = rest[:j]
		}
		repo = host + rest
	} else if j := strings.Index(repo, ":"); j >= 0 {
		repo = repo[:j]
	}
	return repo
}
