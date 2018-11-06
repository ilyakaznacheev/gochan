package gochan

import (
	"context"
)

// Resolver resolvers GraphQL requests
type Resolver struct {
	rh *repoHandler
}

func newResolver(rh *repoHandler) *Resolver {
	return &Resolver{rh}
}

// GetHome resolves getHome query
func (r *Resolver) GetHome(ctx context.Context) (*HomeReprGQL, error) {
	return &HomeReprGQL{}, nil
}

// GetBoard resolves getBoard query
func (r *Resolver) GetBoard(ctx context.Context, args struct{ ID string }) (*BoardReprGQL, error) {
	return &BoardReprGQL{}, nil
}

// GetThread resolves getThread query
func (r *Resolver) GetThread(ctx context.Context, args struct{ ID int32 }) (*ThreadReprGQL, error) {
	return &ThreadReprGQL{}, nil
}

// GetPost resolves getPost query
func (r *Resolver) GetPost(ctx context.Context, args struct{ ID int32 }) (*PostReprGQL, error) {
	return &PostReprGQL{}, nil
}

// GetAuthor resolves getAuthor query
func (r *Resolver) GetAuthor(ctx context.Context, args struct{ ID string }) (*AuthorReprGQL, error) {
	return &AuthorReprGQL{}, nil
}

func (r *Resolver) addPost(ctx context.Context, args struct {
	BoardID string
	Post    PostInputGQL
}) (
	*PostReprGQL, error,
) {
	return &PostReprGQL{}, nil
}

// AddPost resolves addPost mutation
func (r *Resolver) AddPost(ctx context.Context, args struct {
	ThreadID int32
	Post     PostInputGQL
}) (
	*PostReprGQL, error,
) {
	return &PostReprGQL{}, nil
}

// AddThread resolves addThread mutation
func (r *Resolver) AddThread(ctx context.Context, args struct {
	BoardID string
	Thread  ThreadInputGQL
}) (
	*ThreadReprGQL, error,
) {
	return &ThreadReprGQL{}, nil
}
