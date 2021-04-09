package api

import (
	"context"
	"fmt"
	"time"
)

// UsersService provides access to the search related functions
// in the Sumo Logic API.
//
// Sumo Logic API docs: INDECENT EXPOSURE IN PUBLIC
type UsersService service

func (s *UsersService) Users(ctx context.Context, limit int32) (*UsersResult, *Response, error) {

	url := fmt.Sprintf("v1/users?limit=%d", limit)
	req, err := s.client.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, err
	}

	result := new(UsersResult)
	resp, err := s.client.Do(ctx, req, &result)
	if err != nil {
		return nil, resp, err
	}

	return result, resp, nil
}

type UserModel struct {
	// First name of the user.
	FirstName string `json:"firstName"`
	// Last name of the user.
	LastName string `json:"lastName"`
	// Email address of the user.
	Email string `json:"email"`
	// List of roleIds associated with the user.
	RoleIds []string `json:"roleIds"`
	// Creation timestamp in UTC in [RFC3339](https://tools.ietf.org/html/rfc3339) format.
	CreatedAt time.Time `json:"createdAt"`
	// Identifier of the user who created the resource.
	CreatedBy string `json:"createdBy"`
	// Last modification timestamp in UTC.
	ModifiedAt time.Time `json:"modifiedAt"`
	// Identifier of the user who last modified the resource.
	ModifiedBy string `json:"modifiedBy"`
	// Unique identifier for the user.
	Id string `json:"id"`
	// True if the user is active.
	IsActive bool `json:"isActive,omitempty"`
	// This has the value `true` if the user's account has been locked. If a user tries to log into their account several times and fails, his or her account will be locked for security reasons.
	IsLocked bool `json:"isLocked,omitempty"`
	// True if multi factor authentication is enabled for the user.
	IsMfaEnabled bool `json:"isMfaEnabled,omitempty"`
	// Timestamp of the last login for the user in UTC. Will be null if the user has never logged in.
	LastLoginTimestamp time.Time `json:"lastLoginTimestamp,omitempty"`
}

type UsersResult struct {
	// List of users.
	Data []UserModel `json:"data"`
	// Next continuation token.
	Next string `json:"next,omitempty"`
}
