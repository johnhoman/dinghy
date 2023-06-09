package kustomize

import (
	"fmt"
	"github.com/pkg/errors"
	"io"
)

type errFieldPathSyntax struct {
	ch  byte
	pos int
}

func (err *errFieldPathSyntax) Error() string {
	return fmt.Sprintf("unexpected character: %q (pos %d)", string(err.ch), err.pos)
}

func newErrFieldPathSyntax(ch byte, pos int) error { return &errFieldPathSyntax{pos: pos, ch: ch} }

type fieldPathTokenKind string

const (
	fieldPathTokenKindField       fieldPathTokenKind = "string"
	fieldPathTokenKindIndex       fieldPathTokenKind = "int"
	fieldPathTokenKindFieldSelect fieldPathTokenKind = "select"
)

func newFieldPathTokenFromString(kind fieldPathTokenKind, s string) *fieldPathToken {
	return &fieldPathToken{kind: kind, token: s}
}

func newFieldPathToken(kind fieldPathTokenKind, b ...byte) *fieldPathToken {
	return newFieldPathTokenFromString(kind, string(b))
}

// fieldPathToken represents a field of a field path
type fieldPathToken struct {
	kind       fieldPathTokenKind
	token      string
	itemSelect fieldPathTokenSelect
}

type fieldPathTokenSelect struct {
	field *fieldPathToken
	value *fieldPathToken
}

func newFieldPathLexer(in string) *fieldPathLexer {
	l := &fieldPathLexer{data: in}
	l.readChar()
	return l
}

// fieldPathLexer parses a field path parts into tokens
type fieldPathLexer struct {
	data         string
	position     int
	nextPosition int
	ch           byte
}

// readChar sets the next character from the
// provided string
func (l *fieldPathLexer) readChar() {
	if l.nextPosition >= len(l.data) {
		l.ch = 0
	} else {
		l.ch = l.data[l.nextPosition]
	}
	// set the current position
	l.position = l.nextPosition
	// point to the next token
	l.nextPosition += 1
}

func (l *fieldPathLexer) readWhen(pred func(byte) bool) string {
	if l.position >= len(l.data) {
		// EOF
		return string(byte(0))
	}
	position := l.position

	for pred(l.ch) {
		l.readChar()
	}
	return l.data[position:l.position]
}

func (l *fieldPathLexer) readInt() string {
	return l.readWhen(fieldPathIsNumber)
}

func (l *fieldPathLexer) readString() string {
	return l.readWhen(fieldPathIsLetter)
}

// peekChar returns the next character, but doesn't
// change the state of the lexer.
func (l *fieldPathLexer) peekChar() byte {
	if l.nextPosition >= len(l.data) {
		return 0
	}
	return l.data[l.nextPosition]
}

func (l *fieldPathLexer) nextToken() (*fieldPathToken, error) {

	var token *fieldPathToken

	switch l.ch {
	case '\'', '"':
		var (
			err  error
			open = l.ch
		)
		l.readChar()
		token, err = l.nextToken()
		if err != nil {
			return nil, err
		}
		// if the token in enclosed in quotes, then it
		// should be a field, but integers will be read
		// as integers in l.nextToken because it doesn't
		// know about the quotes
		token.kind = fieldPathTokenKindField
		if l.ch != open {
			return nil, errors.Wrapf(newErrFieldPathSyntax(l.ch, l.position), "expected %q", string(open))
		}
		l.readChar()
	case ']':
		return nil, newErrFieldPathSyntax(l.ch, l.position)
	case '[':
		// token can be a comparison, field or index.
		l.readChar()
		var err error
		token, err = l.nextToken()
		if err != nil {
			return nil, err
		}
		// brackets can contain select statements for narrowing
		// in on specific items in a map. If the next character
		// is an '=', assume it's a select statement
		if l.ch == '=' {
			l.readChar()
			pos := l.position
			ch := l.ch
			value, err := l.nextToken()
			if err != nil {
				return nil, err
			}
			if value.kind != fieldPathTokenKindField {
				return nil, errors.Wrapf(
					newErrFieldPathSyntax(ch, pos),
					"expected %q, got %q",
					fieldPathTokenKindField,
					value.kind,
				)
			}
			token.token = ""
			token.kind = fieldPathTokenKindFieldSelect
			token.itemSelect = fieldPathTokenSelect{
				field: &fieldPathToken{
					kind:  token.kind,
					token: token.token,
				},
				value: &fieldPathToken{
					kind:  value.kind,
					token: value.token,
				},
			}
		}
		if l.ch != ']' {
			return nil, errors.Wrapf(newErrFieldPathSyntax(l.ch, l.position), "expected %q", "]")
		}
		l.readChar()
	case '.':
		l.readChar()
		pos := l.position
		ch := l.ch
		var err error
		token, err = l.nextToken()
		if err != nil {
			return nil, err
		}
		if token.kind != fieldPathTokenKindField {
			return nil, errors.Wrapf(
				newErrFieldPathSyntax(ch, pos),
				"got %s, want %s",
				token.kind,
				fieldPathTokenKindField,
			)
		}
	case 0:
		return nil, io.EOF
	default:
		switch {
		case fieldPathIsNumber(l.ch):
			token = newFieldPathTokenFromString(fieldPathTokenKindIndex, l.readInt())
		case fieldPathIsLetter(l.ch):
			token = newFieldPathTokenFromString(fieldPathTokenKindField, l.readString())
		default:
			return nil, newErrFieldPathSyntax(l.ch, l.position)
		}
	}
	return token, nil
}
func (l *fieldPathLexer) parseTokens() ([]fieldPathToken, error) {
	rv := make([]fieldPathToken, 0)
	for {
		token, err := l.nextToken()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		rv = append(rv, *token)
	}
	return rv, nil
}

func fieldPathIsNumber(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func fieldPathIsLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z'
}
