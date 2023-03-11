package dejavu

import (
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type PlaceholderSyntax string

const (
	PlaceholderIndexed      = "Indexed"
	PlaceholderNamed        = "Named"
	PlaceholderQuestionMark = "Question Mark"
)

type Statement struct {
	sql  string
	args []sql.NamedArg
}

func NewStatement(format string, args ...any) *Statement {
	return &Statement{sql: fmt.Sprintf(format, args...)}
}

func (s *Statement) Arg(name string, value any) *Statement {
	s.args = append(s.args, sql.Named(name, value))

	return s
}

func (s *Statement) WithSyntax(syntax PlaceholderSyntax) (string, []any) {
	switch syntax {
	case PlaceholderIndexed:
		return s.WithIndexedArgs()
	case PlaceholderNamed:
		return s.WithNamedArgs()
	case PlaceholderQuestionMark:
		return s.WithQuestionMarkArgs()
	}

	panic("don't know how to handle " + syntax)
}

func (s *Statement) WithIndexedArgs() (string, []any) {
	args := make([]any, len(s.args))
	stmt := s.sql

	for i, arg := range s.args {
		stmt = strings.ReplaceAll(stmt, ":"+arg.Name, "$"+strconv.Itoa(i+1))
		args[i] = arg.Value
	}

	return stmt, args
}

func (s *Statement) WithNamedArgs() (string, []any) {
	args := make([]any, len(s.args))

	for i, arg := range s.args {
		args[i] = arg
	}

	return s.sql, args
}

type argIndex struct {
	idx   int
	value any
}

func (s *Statement) WithQuestionMarkArgs() (string, []any) {
	argIndexes := make([]argIndex, 0, len(s.args))
	stmt := s.sql

	for _, arg := range s.args {
		for _, idx := range allIndexes(s.sql, ":"+arg.Name) {
			argIndexes = append(argIndexes, argIndex{
				idx:   idx,
				value: arg.Value,
			})
		}

		stmt = strings.ReplaceAll(stmt, ":"+arg.Name, "?")
	}

	sort.Slice(argIndexes, func(i, j int) bool {
		return argIndexes[i].idx < argIndexes[j].idx
	})

	args := make([]any, len(argIndexes))

	for i, arg := range argIndexes {
		args[i] = arg.value
	}

	return stmt, args
}

func (s *Statement) String() string {
	sb := strings.Builder{}

	sb.WriteString(s.sql)

	for _, arg := range s.args {
		sb.WriteRune('\n')
		sb.WriteString(fmt.Sprintf(" -- %s=%v", arg.Name, arg.Value))
	}

	return sb.String()
}

func allIndexes(s, substr string) []int {
	result := make([]int, 0, 1)

	for remaining, prefixLen := s, 0; ; {
		i := strings.Index(remaining, substr)

		if i == -1 {
			break
		}

		result = append(result, prefixLen+i)
		prefixLen += i + len(substr)
		remaining = remaining[i+len(substr):]
	}

	return result
}
