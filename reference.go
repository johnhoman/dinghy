package kustomize

import "regexp"

type ReferenceType string

const (
	// ReferenceTypeLocal is a reference to a file in the current build
	// path. If the current build path is a GitHub repo, this will be a
	// file in the same GitHub repo. If the current build path if local,
	// this will be a path in the local file system.
	ReferenceTypeLocal ReferenceType = "Local"

	// ReferenceTypeRemote is a reference to a build path outside the
	// current build path. Currently, only GitHub is supported
	ReferenceTypeRemote ReferenceType = "Remote"

	// ReferenceTypeRemoteGitHub is a reference to a GitHub source
	// build path. GitHub will only be supported over https, so a valid
	// PAT will be required
	ReferenceTypeRemoteGitHub ReferenceType = "Remote:GitHub"

	// ReferenceTypeRemoteS3 is a reference to an S3 source build path.
	ReferenceTypeRemoteS3 ReferenceType = "Remote:S3"

	// ReferenceTypeUnknown means that the provided reference wasn't
	// resolved by the parser.
	ReferenceTypeUnknown ReferenceType = "Unknown"
)

// ParseReferenceType parses a path reference from the provided
// reference to find the origin, which will be a variant of
// either local or remote.
func ParseReferenceType(path string) ReferenceType {
	switch {
	case git.MatchString(path):
		return ReferenceTypeRemoteGitHub
	case s3.MatchString(path):
		return ReferenceTypeRemoteS3
	case remote.MatchString(path):
		return ReferenceTypeRemote
	}
	return ReferenceTypeLocal
}

var (
	git    = regexp.MustCompile(`^(?:https:\/\/)?(?:www\.)?github\.com\b(?:[-a-zA-Z0-9()@:%_\+.~#?&\/=]*)$`)
	s3     = regexp.MustCompile(`^s3:\/\/[-a-z0-9@:%._\+~#=]{1,256}\b(?:[-a-zA-Z0-9()@:%_\+.~#?&\/=]*)$`)
	remote = regexp.MustCompile(`^https?:\/\/(?:www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b(?:[-a-zA-Z0-9()@:%_\+.~#?&\/=]*)$`)
)
