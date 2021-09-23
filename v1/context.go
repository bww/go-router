package router

import (
	"context"
)

const matchKey = "github.com/bww/go-router.Match"

func newMatchContext(cxt context.Context, match *Match) context.Context {
	return context.WithValue(cxt, matchKey, match)
}

func MatchFromContext(cxt context.Context) *Match {
	match, ok := cxt.Value(matchKey).(*Match)
	if ok {
		return match
	} else {
		return nil
	}
}
