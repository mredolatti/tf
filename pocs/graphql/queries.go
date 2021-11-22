package main

import (
	"fmt"

	"github.com/graphql-go/graphql"
)

type QueryProcessor struct {
	schema graphql.Schema
}

func NewProcessor() (*QueryProcessor, error) {
	fields := graphql.Fields{
		"hello": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				fmt.Printf("%+v\n", p)
				return "world", nil
			},
		},
	}
	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		return nil, fmt.Errorf("error instantiating GraphQL Schema: %w", err)
	}

	return &QueryProcessor{schema: schema}, nil

}

func (p *QueryProcessor) Handle(query string) (*graphql.Result, error) {
	params := graphql.Params{Schema: p.schema, RequestString: query}
	return graphql.Do(params), nil
}
