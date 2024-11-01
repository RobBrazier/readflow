// Code generated by github.com/Khan/genqlient, DO NOT EDIT.

package anilist

import (
	"context"

	"github.com/Khan/genqlient/graphql"
)

// GetCurrentUserResponse is returned by GetCurrentUser on success.
type GetCurrentUserResponse struct {
	// Get the currently authenticated user
	Viewer GetCurrentUserViewerUser `json:"Viewer"`
}

// GetViewer returns GetCurrentUserResponse.Viewer, and is useful for accessing the field via an interface.
func (v *GetCurrentUserResponse) GetViewer() GetCurrentUserViewerUser { return v.Viewer }

// GetCurrentUserViewerUser includes the requested fields of the GraphQL type User.
// The GraphQL type's documentation follows.
//
// A user
type GetCurrentUserViewerUser struct {
	// The id of the user
	Id int `json:"id"`
	// The name of the user
	Name string `json:"name"`
}

// GetId returns GetCurrentUserViewerUser.Id, and is useful for accessing the field via an interface.
func (v *GetCurrentUserViewerUser) GetId() int { return v.Id }

// GetName returns GetCurrentUserViewerUser.Name, and is useful for accessing the field via an interface.
func (v *GetCurrentUserViewerUser) GetName() string { return v.Name }

// The query or mutation executed by GetCurrentUser.
const GetCurrentUser_Operation = `
query GetCurrentUser {
	Viewer {
		id
		name
	}
}
`

func GetCurrentUser(
	ctx_ context.Context,
	client_ graphql.Client,
) (*GetCurrentUserResponse, error) {
	req_ := &graphql.Request{
		OpName: "GetCurrentUser",
		Query:  GetCurrentUser_Operation,
	}
	var err_ error

	var data_ GetCurrentUserResponse
	resp_ := &graphql.Response{Data: &data_}

	err_ = client_.MakeRequest(
		ctx_,
		req_,
		resp_,
	)

	return &data_, err_
}
