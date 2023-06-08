package kustomize

import (
	qt "github.com/frankban/quicktest"
	"testing"
)

func TestParseReference(t *testing.T) {
	cases := map[string]ReferenceType{
		"https://github.com/john-homan/nop.git":                               ReferenceTypeRemoteGitHub,
		"https://github.com/john-homan/nop.git/path/to/file":                  ReferenceTypeRemoteGitHub,
		"https://github.com/john-homan/nop.git/path/to/file?ref=asdfasdfasdf": ReferenceTypeRemoteGitHub,
		"https://github.com/john.homan/nop.git":                               ReferenceTypeRemoteGitHub,
		"https://github.com/john.homan/nop.git/path/to/file":                  ReferenceTypeRemoteGitHub,
		"https://github.com/john.homan/nop.git/path/to/file?ref=asdfasdfasdf": ReferenceTypeRemoteGitHub,
		"https://github.com/johnhoman/nop.git":                                ReferenceTypeRemoteGitHub,
		"https://github.com/johnhoman/nop.git/path/to/file":                   ReferenceTypeRemoteGitHub,
		"https://github.com/johnhoman/nop.git/path/to/file?ref=asdfasdfasdf":  ReferenceTypeRemoteGitHub,
		"https://github.com/johnhoman/nop":                                    ReferenceTypeRemoteGitHub,
		"https://github.com/johnhoman/nop/path/to/file":                       ReferenceTypeRemoteGitHub,
		"https://github.com/johnhoman/nop/path/to/file?ref=asdfasdfasdf":      ReferenceTypeRemoteGitHub,
		"github.com/johnhoman/nop":                                            ReferenceTypeRemoteGitHub,
		"github.com/johnhoman/nop/path/to/file":                               ReferenceTypeRemoteGitHub,
		"github.com/johnhoman/nop/path/to/file?ref=asdfasdfasdf":              ReferenceTypeRemoteGitHub,
		"github.com/johnhoman/nop.git":                                        ReferenceTypeRemoteGitHub,
		"github.com/johnhoman/nop.git/path/to/file":                           ReferenceTypeRemoteGitHub,
		"github.com/johnhoman/nop.git/path/to/file?ref=asdfasdfasdf":          ReferenceTypeRemoteGitHub,
		"https://s3.com/johnhoman/nop":                                        ReferenceTypeRemote,
		"https://s3.com/johnhoman/nop/path/to/file":                           ReferenceTypeRemote,
		"https://s3.com/johnhoman/nop/path/to/file?ref=asdfasdfasdf":          ReferenceTypeRemote,
		"s3://my-bucket/path/to/key":                                          ReferenceTypeRemoteS3,
	}

	for ref, typ := range cases {
		t.Run(ref, func(t *testing.T) {
			qt.Assert(t, ParseReferenceType(ref), qt.Equals, typ)
		})
	}
}
