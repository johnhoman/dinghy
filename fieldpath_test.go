package kustomize

import (
	qt "github.com/frankban/quicktest"
	"io"
	"testing"
)

func TestFieldPathLexer(t *testing.T) {
	cases := map[string][]*fieldPathToken{
		"data.FOO": {
			{kind: fieldPathTokenKindField, token: "data"},
			{kind: fieldPathTokenKindField, token: "FOO"},
		},
		"data[FOO]": {
			{kind: fieldPathTokenKindField, token: "data"},
			{kind: fieldPathTokenKindField, token: "FOO"},
		},
		"data['FOO']": {
			{kind: fieldPathTokenKindField, token: "data"},
			{kind: fieldPathTokenKindField, token: "FOO"},
		},
		`data["FOO"]`: {
			{kind: fieldPathTokenKindField, token: "data"},
			{kind: fieldPathTokenKindField, token: "FOO"},
		},
		`data["FOO"].bar`: {
			{kind: fieldPathTokenKindField, token: "data"},
			{kind: fieldPathTokenKindField, token: "FOO"},
			{kind: fieldPathTokenKindField, token: "bar"},
		},
		`data["FOO"].bar[0]`: {
			{kind: fieldPathTokenKindField, token: "data"},
			{kind: fieldPathTokenKindField, token: "FOO"},
			{kind: fieldPathTokenKindField, token: "bar"},
			{kind: fieldPathTokenKindIndex, token: "0"},
		},
		`data[0].bar`: {
			{kind: fieldPathTokenKindField, token: "data"},
			{kind: fieldPathTokenKindIndex, token: "0"},
			{kind: fieldPathTokenKindField, token: "bar"},
		},
		`data["0"].bar`: {
			{kind: fieldPathTokenKindField, token: "data"},
			{kind: fieldPathTokenKindField, token: "0"},
			{kind: fieldPathTokenKindField, token: "bar"},
		},
		`data[name=bar].bar`: {
			{kind: fieldPathTokenKindField, token: "data"},
			{kind: fieldPathTokenKindFieldSelect, itemSelect: fieldPathTokenSelect{
				field: &fieldPathToken{
					kind:  fieldPathTokenKindField,
					token: "name",
				},
				value: &fieldPathToken{
					kind:  fieldPathTokenKindField,
					token: "bar",
				},
			}},
			{kind: fieldPathTokenKindField, token: "bar"},
		},
		`data[name="bar"].bar`: {
			{kind: fieldPathTokenKindField, token: "data"},
			{kind: fieldPathTokenKindFieldSelect, itemSelect: fieldPathTokenSelect{
				field: &fieldPathToken{
					kind:  fieldPathTokenKindField,
					token: "name",
				},
				value: &fieldPathToken{
					kind:  fieldPathTokenKindField,
					token: "bar",
				},
			}},
			{kind: fieldPathTokenKindField, token: "bar"},
		},
		`data[name='bar'].bar`: {
			{kind: fieldPathTokenKindField, token: "data"},
			{kind: fieldPathTokenKindFieldSelect, itemSelect: fieldPathTokenSelect{
				field: &fieldPathToken{
					kind:  fieldPathTokenKindField,
					token: "name",
				},
				value: &fieldPathToken{
					kind:  fieldPathTokenKindField,
					token: "bar",
				},
			}},
			{kind: fieldPathTokenKindField, token: "bar"},
		},
	}
	for fieldPath, tokens := range cases {
		t.Run(fieldPath, func(t *testing.T) {
			l := newFieldPathLexer(fieldPath)
			for _, token := range tokens {
				next, err := l.nextToken()
				qt.Assert(t, err, qt.IsNil)
				qt.Assert(t, next.kind, qt.Equals, token.kind)
				qt.Assert(t, next.token, qt.Equals, token.token)
			}
			_, err := l.nextToken()
			qt.Assert(t, err, qt.ErrorIs, io.EOF)
		})
	}
}

func TestFieldPathLexer_Error(t *testing.T) {
	cases := map[string]struct {
		pos int
		ch  byte
	}{
		"data.FOO]":         {pos: 8, ch: ']'},
		"data.[FOO":         {pos: 9, ch: 0}, // EOF
		"data.0.bar":        {pos: 5, ch: '0'},
		`data.["0].bar`:     {pos: 8, ch: ']'},  // expected "
		`data.['0].bar`:     {pos: 8, ch: ']'},  // expected '
		`data.[0'].bar`:     {pos: 7, ch: '\''}, // expected ]
		`data.[0"].bar`:     {pos: 7, ch: '"'},
		`data.[name=0].bar`: {pos: 11, ch: '0'},
	}
	for fieldPath, subtest := range cases {
		t.Run(fieldPath, func(t *testing.T) {
			l := newFieldPathLexer(fieldPath)
			_, err := l.parseTokens()
			syntaxErr := &errFieldPathSyntax{}
			qt.Assert(t, err, qt.ErrorAs, &syntaxErr)
			qt.Assert(t, syntaxErr.pos, qt.Equals, subtest.pos)
			qt.Assert(t, string(syntaxErr.ch), qt.Equals, string(subtest.ch))
		})
	}
}
