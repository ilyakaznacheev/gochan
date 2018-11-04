package main

import (
	"context"
	"io/ioutil"

	"github.com/graph-gophers/graphql-go"
)

func getSchema(filename string, rh *repoHandler) (*graphql.Schema, error) {
	schemaFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	schemaRaw := string(schemaFile)

	return graphql.MustParseSchema(schemaRaw, newResolver(rh)), nil
}

// Resolver types

// HomeReprGQL is GQL Home representation structure
type HomeReprGQL struct {
	// Boards []*BoardReprGQL
}

// BOARDS field
func (r *HomeReprGQL) BOARDS(ctx context.Context) *[]*BoardReprGQL {
	return &[]*BoardReprGQL{}
}

// BoardReprGQL is GQL Board representation structure
type BoardReprGQL struct {
	// ID      string
	// Title   string
	// Threads []*ThreadReprGQL
}

func (r *BoardReprGQL) ID(ctx context.Context) *string {
	res := ""
	return &res
}

func (r *BoardReprGQL) TITLE(ctx context.Context) *string {
	res := ""
	return &res
}

func (r *BoardReprGQL) THREADS(ctx context.Context) *[]*ThreadReprGQL {
	return &[]*ThreadReprGQL{}
}

// ThreadReprGQL is GQL Thread representation structure
type ThreadReprGQL struct {
	// ID     string
	// Title  string
	// Head   *PostReprGQL
	// Posts  []*PostReprGQL
	// Author *AuthorReprGQL
}

func (r *ThreadReprGQL) ID(ctx context.Context) *graphql.ID {
	res := graphql.ID("")
	return &res
}

func (r *ThreadReprGQL) TITLE(ctx context.Context) *string {
	res := ""
	return &res
}

func (r *ThreadReprGQL) HEAD(ctx context.Context) *PostReprGQL {
	return &PostReprGQL{}
}

func (r *ThreadReprGQL) POSTS(ctx context.Context) *[]*PostReprGQL {
	return &[]*PostReprGQL{}
}

func (r *ThreadReprGQL) AUTHOR(ctx context.Context) *AuthorReprGQL {
	return &AuthorReprGQL{}
}

// PostReprGQL is GQL Post representation structure
type PostReprGQL struct {
	// ID    string
	// Text  string
	// Img   *ImageReprGQL
}

func (r *PostReprGQL) ID(ctx context.Context) *graphql.ID {
	res := graphql.ID("")
	return &res
}

func (r *PostReprGQL) TEXT(ctx context.Context) *string {
	res := ""
	return &res
}

func (r *PostReprGQL) IMG(ctx context.Context) *ImageReprGQL {
	return &ImageReprGQL{}
}

func (r *PostReprGQL) AUTHOR(ctx context.Context) *AuthorReprGQL {
	return &AuthorReprGQL{}
}

// AuthorReprGQL is GQL Author representation structure
type AuthorReprGQL struct {
	// ID    string
	// Posts []*PostReprGQL
}

func (r *AuthorReprGQL) ID(ctx context.Context) *string {
	res := ""
	return &res
}

func (r *AuthorReprGQL) POSTS(ctx context.Context) *[]*PostReprGQL {
	return &[]*PostReprGQL{}
}

// ImageReprGQL is GQL Image representation structure
type ImageReprGQL struct {
	// URL string
}

func (r *ImageReprGQL) URL(ctx context.Context) *string {
	res := ""
	return &res
}

// PostInputGQL is GQL Post input structure
type PostInputGQL struct {
	// Title string
	// Text  string
}

// ThreadInputGQL is GQL Thread input structure
type ThreadInputGQL struct {
	// Post PostInputGQL
}
