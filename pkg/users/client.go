package users

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-querystring/query"
	"github.com/workos/workos-go/v2/internal/workos"
	"github.com/workos/workos-go/v2/pkg/common"
	"github.com/workos/workos-go/v2/pkg/workos_errors"
	"net/http"
	"time"
)

// ResponseLimit is the default number of records to limit a response to.
const ResponseLimit = 10

// Order represents the order of records.
type Order string

// Constants that enumerate the available orders.
const (
	Asc  Order = "asc"
	Desc Order = "desc"
)

// Organization contains data about a particular Organization.
type Organization struct {
	// The Organization's unique identifier.
	ID string `json:"id"`

	// The Organization's name.
	Name string `json:"name"`
}

// OrganizationMembership contains data about a particular OrganizationMembership.
type OrganizationMembership struct {
	// Contains the ID and name of the associated Organization.
	Organization Organization `json:"organization"`

	// CreatedAt is the timestamp of when the OrganizationMembership was created.
	CreatedAt string `json:"created_at"`

	// UpdatedAt is the timestamp of when the OrganizationMembership was updated.
	UpdatedAt string `json:"updated_at"`
}

// User contains data about a particular User.
type User struct {

	// The User's unique identifier.
	ID string `json:"id"`

	// The User's first name.
	FirstName string `json:"first_name"`

	// The User's last name.
	LastName string `json:"last_name"`

	// The User's email.
	Email string `json:"email"`

	// The timestamp of when the User was created.
	CreatedAt string `json:"created_at"`

	// The timestamp of when the User was updated.
	UpdatedAt string `json:"updated_at"`

	// Whether the User email is verified.
	EmailVerified bool `json:"email_verified"`
}

// GetUserOpts contains the options to pass in order to get a user profile.
type GetUserOpts struct {
	// User unique identifier
	User string `json:"id"`
}

// ListUsersResponse contains the response from the ListUsers call.
type ListUsersResponse struct {
	// List of Users
	Data []User `json:"data"`

	// Cursor to paginate through the list of Users
	ListMetadata common.ListMetadata `json:"listMetadata"`
}

type ListUsersOpts struct {
	// Filter Users by their email.
	Email string `url:"email,omitempty"`

	// Filter Users by the organization they are members of.
	Organization string `url:"organization,omitempty"`

	// Maximum number of records to return.
	Limit int `url:"limit"`

	// The order in which to paginate records.
	Order Order `url:"order,omitempty"`

	// Pagination cursor to receive records before a provided User ID.
	Before string `url:"before,omitempty"`

	// Pagination cursor to receive records after a provided User ID.
	After string `url:"after,omitempty"`
}

type CreateUserOpts struct {
	Email         string `json:"email"`
	Password      string `json:"password,omitempty"`
	FirstName     string `json:"first_name,omitempty"`
	LastName      string `json:"last_name,omitempty"`
	EmailVerified bool   `json:"email_verified,omitempty"`
}

type UpdateUserOpts struct {
	User          string
	FirstName     string `json:"first_name,omitempty"`
	LastName      string `json:"last_name,omitempty"`
	EmailVerified bool   `json:"email_verified,omitempty"`
}

type UpdateUserPasswordOpts struct {
	User     string
	Password string `json:"password"`
}

type DeleteUserOpts struct {
	User string
}

type AuthorizedOrganization struct {
	Organization Organization `json:"organization"`
}

type AuthenticateWithPasswordOpts struct {
	ClientID  string `json:"client_id"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

type AuthenticateWithCodeOpts struct {
	ClientID  string `json:"client_id"`
	Code      string `json:"code"`
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

type MagicAuthChallengeID string

type AuthenticateWithMagicAuthOpts struct {
	ClientID  string `json:"client_id"`
	Code      string `json:"code"`
	User      string `json:"user_id"`
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

type AuthenticationResponse struct {
	User User `json:"user"`
}

type SendVerificationEmailOpts struct {
	// The unique ID of the User whose email address will be verified.
	User string
}

type VerifyEmailCodeOpts struct {
	// The unique ID of the User whose code will be verified.
	User string
	// The verification code emailed to the user.
	Code string `json:"code"`
}

type CreatePasswordResetChallengeOpts struct {
	// The unique ID of the User whose email address will be verified.
	Email string `json:"email"`

	// The URL that will be linked to in the verification email.
	PasswordResetUrl string `json:"password_reset_url"`
}

type CompletePasswordResetOpts struct {
	// The verification token emailed to the user.
	Token string `json:"token"`

	// The new password to be set for the user.
	NewPassword string `json:"new_password"`
}

type UserResponse struct {
	User User `json:"user"`
}

type SendMagicAuthCodeOpts struct {
	// The email address the one-time code will be sent to.
	Email string `json:"email_address"`
}

type AddUserToOrganizationOpts struct {
	User         string `json:"id"`
	Organization string `json:"organization_id"`
}

type RemoveUserFromOrganizationOpts struct {
	User         string `json:"id"`
	Organization string `json:"organization_id"`
}

func NewClient(apiKey string) *Client {
	return &Client{
		APIKey:     apiKey,
		Endpoint:   "https://api.workos.com",
		HTTPClient: &http.Client{Timeout: time.Second * 10},
		JSONEncode: json.Marshal,
	}
}

// GetUser returns details of an existing user
func (c *Client) GetUser(ctx context.Context, opts GetUserOpts) (User, error) {
	endpoint := fmt.Sprintf(
		"%s/users/%s",
		c.Endpoint,
		opts.User,
	)

	req, err := http.NewRequest(
		http.MethodGet,
		endpoint,
		nil,
	)
	if err != nil {
		return User{}, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("User-Agent", "workos-go/"+workos.Version)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return User{}, err
	}
	defer res.Body.Close()

	if err = workos_errors.TryGetHTTPError(res); err != nil {
		return User{}, err
	}

	var body User
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&body)

	return body, err
}

// ListUsers get a list of all of your existing users matching the criteria specified.
func (c *Client) ListUsers(ctx context.Context, opts ListUsersOpts) (ListUsersResponse, error) {
	endpoint := fmt.Sprintf(
		"%s/users",
		c.Endpoint,
	)

	req, err := http.NewRequest(
		http.MethodGet,
		endpoint,
		nil,
	)
	if err != nil {
		return ListUsersResponse{}, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("User-Agent", "workos-go/"+workos.Version)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	if opts.Limit == 0 {
		opts.Limit = ResponseLimit
	}

	queryValues, err := query.Values(opts)
	if err != nil {
		return ListUsersResponse{}, err
	}

	req.URL.RawQuery = queryValues.Encode()

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return ListUsersResponse{}, err
	}
	defer res.Body.Close()

	if err = workos_errors.TryGetHTTPError(res); err != nil {
		return ListUsersResponse{}, err
	}

	var body ListUsersResponse
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&body)

	return body, err
}

// CreateUser create a new user with email password authentication.
// Only unmanaged users can be created directly using the User Management API.
func (c *Client) CreateUser(ctx context.Context, opts CreateUserOpts) (User, error) {
	endpoint := fmt.Sprintf(
		"%s/users",
		c.Endpoint,
	)

	data, err := c.JSONEncode(opts)
	if err != nil {
		return User{}, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		endpoint,
		bytes.NewBuffer(data),
	)
	if err != nil {
		return User{}, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("User-Agent", "workos-go/"+workos.Version)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return User{}, err
	}
	defer res.Body.Close()

	if err = workos_errors.TryGetHTTPError(res); err != nil {
		return User{}, err
	}

	var body User
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&body)

	return body, err
}

// UpdateUser updates User attributes.
func (c *Client) UpdateUser(ctx context.Context, opts UpdateUserOpts) (User, error) {
	endpoint := fmt.Sprintf(
		"%s/users/%s",
		c.Endpoint,
		opts.User,
	)

	data, err := c.JSONEncode(opts)
	if err != nil {
		return User{}, err
	}

	req, err := http.NewRequest(
		http.MethodPut,
		endpoint,
		bytes.NewBuffer(data),
	)
	if err != nil {
		return User{}, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("User-Agent", "workos-go/"+workos.Version)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return User{}, err
	}
	defer res.Body.Close()

	if err = workos_errors.TryGetHTTPError(res); err != nil {
		return User{}, err
	}

	var body User
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&body)

	return body, err
}

// UpdateUserPassword updates a User password.
func (c *Client) UpdateUserPassword(ctx context.Context, opts UpdateUserPasswordOpts) (User, error) {
	endpoint := fmt.Sprintf(
		"%s/users/%s/password",
		c.Endpoint,
		opts.User,
	)

	data, err := c.JSONEncode(opts)
	if err != nil {
		return User{}, err
	}

	req, err := http.NewRequest(
		http.MethodPut,
		endpoint,
		bytes.NewBuffer(data),
	)
	if err != nil {
		return User{}, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("User-Agent", "workos-go/"+workos.Version)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return User{}, err
	}
	defer res.Body.Close()

	if err = workos_errors.TryGetHTTPError(res); err != nil {
		return User{}, err
	}

	var body User
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&body)

	return body, err
}

// DeleteUser delete an existing user.
func (c *Client) DeleteUser(ctx context.Context, opts DeleteUserOpts) error {
	endpoint := fmt.Sprintf(
		"%s/users/%s",
		c.Endpoint,
		opts.User,
	)

	req, err := http.NewRequest(
		http.MethodDelete,
		endpoint,
		nil,
	)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	req.Header.Set("User-Agent", "workos-go/"+workos.Version)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return workos_errors.TryGetHTTPError(res)
}

// AddUserToOrganization adds an unmanaged user to an Organization
func (c *Client) AddUserToOrganization(ctx context.Context, opts AddUserToOrganizationOpts) (User, error) {
	endpoint := fmt.Sprintf(
		"%s/users/%s/organizations",
		c.Endpoint,
		opts.User,
	)

	data, err := c.JSONEncode(opts)
	if err != nil {
		return User{}, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		endpoint,
		bytes.NewBuffer(data),
	)
	if err != nil {
		return User{}, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("User-Agent", "workos-go/"+workos.Version)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return User{}, err
	}
	defer res.Body.Close()

	if err = workos_errors.TryGetHTTPError(res); err != nil {
		return User{}, err
	}

	var body User
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&body)

	return body, err
}

// RemoveUserFromOrganization removes an unmanaged User from the given Organization.
func (c *Client) RemoveUserFromOrganization(ctx context.Context, opts RemoveUserFromOrganizationOpts) (User, error) {
	endpoint := fmt.Sprintf(
		"%s/users/%s/organizations/%s",
		c.Endpoint,
		opts.User,
		opts.Organization,
	)

	req, err := http.NewRequest(
		http.MethodDelete,
		endpoint,
		nil,
	)
	if err != nil {
		return User{}, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("User-Agent", "workos-go/"+workos.Version)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return User{}, err
	}
	defer res.Body.Close()

	if err = workos_errors.TryGetHTTPError(res); err != nil {
		return User{}, err
	}

	var body User
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&body)

	return body, err
}

// AuthenticateWithPassword authenticates a user with Email and Password
func (c *Client) AuthenticateWithPassword(ctx context.Context, opts AuthenticateWithPasswordOpts) (AuthenticationResponse, error) {
	payload := struct {
		AuthenticateWithPasswordOpts
		ClientSecret string `json:"client_secret"`
		GrantType    string `json:"grant_type"`
	}{
		AuthenticateWithPasswordOpts: opts,
		ClientSecret:                 c.APIKey,
		GrantType:                    "password",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return AuthenticationResponse{}, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		c.Endpoint+"/users/authenticate",
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		return AuthenticationResponse{}, err
	}

	// Add headers and context to the request
	req = req.WithContext(ctx)
	req.Header.Set("User-Agent", "workos-go/"+workos.Version)
	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return AuthenticationResponse{}, err
	}
	defer res.Body.Close()

	if err = workos_errors.TryGetHTTPError(res); err != nil {
		return AuthenticationResponse{}, err
	}

	// Parse the JSON response
	var body AuthenticationResponse
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&body)

	return body, err
}

// AuthenticateWithCode authenticates an OAuth user or a managed SSO user that is logging in through SSO
func (c *Client) AuthenticateWithCode(ctx context.Context, opts AuthenticateWithCodeOpts) (AuthenticationResponse, error) {
	payload := struct {
		AuthenticateWithCodeOpts
		ClientSecret string `json:"client_secret"`
		GrantType    string `json:"grant_type"`
	}{
		AuthenticateWithCodeOpts: opts,
		ClientSecret:             c.APIKey,
		GrantType:                "authorization_code",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return AuthenticationResponse{}, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		c.Endpoint+"/users/authenticate",
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		return AuthenticationResponse{}, err
	}

	// Add headers and context to the request
	req = req.WithContext(ctx)
	req.Header.Set("User-Agent", "workos-go/"+workos.Version)
	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return AuthenticationResponse{}, err
	}
	defer res.Body.Close()

	if err = workos_errors.TryGetHTTPError(res); err != nil {
		return AuthenticationResponse{}, err
	}

	// Parse the JSON response
	var body AuthenticationResponse
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&body)

	return body, err
}

// AuthenticateWithMagicAuth authenticates a user by verifying a one-time code sent to the user's email address by
// the Magic Auth Send Code endpoint.
func (c *Client) AuthenticateWithMagicAuth(ctx context.Context, opts AuthenticateWithMagicAuthOpts) (AuthenticationResponse, error) {
	payload := struct {
		AuthenticateWithMagicAuthOpts
		ClientSecret string `json:"client_secret"`
		GrantType    string `json:"grant_type"`
	}{
		AuthenticateWithMagicAuthOpts: opts,
		ClientSecret:                  c.APIKey,
		GrantType:                     "urn:workos:oauth:grant-type:magic-auth:code",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return AuthenticationResponse{}, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		c.Endpoint+"/users/authenticate",
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		return AuthenticationResponse{}, err
	}

	// Add headers and context to the request
	req = req.WithContext(ctx)
	req.Header.Set("User-Agent", "workos-go/"+workos.Version)
	req.Header.Set("Content-Type", "application/json")

	// Execute the request
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return AuthenticationResponse{}, err
	}
	defer res.Body.Close()

	if err = workos_errors.TryGetHTTPError(res); err != nil {
		return AuthenticationResponse{}, err
	}

	// Parse the JSON response
	var body AuthenticationResponse
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&body)

	return body, err
}

// SendVerificationEmail creates an email verification challenge and emails verification token to user.
func (c *Client) SendVerificationEmail(ctx context.Context, opts SendVerificationEmailOpts) (UserResponse, error) {
	endpoint := fmt.Sprintf(
		"%s/users/%s/send_verification_email",
		c.Endpoint,
		opts.User,
	)
	req, err := http.NewRequest(
		http.MethodPost,
		endpoint,
		nil,
	)
	if err != nil {
		return UserResponse{}, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("User-Agent", "workos-go/"+workos.Version)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return UserResponse{}, err
	}
	defer res.Body.Close()

	if err = workos_errors.TryGetHTTPError(res); err != nil {
		return UserResponse{}, err
	}

	var body UserResponse
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&body)

	return body, err
}

// VerifyEmailCode verifies user email using verification token that was sent to the user.
func (c *Client) VerifyEmailCode(ctx context.Context, opts VerifyEmailCodeOpts) (UserResponse, error) {
	endpoint := fmt.Sprintf(
		"%s/users/%s/verify_email_code",
		c.Endpoint,
		opts.User,
	)

	data, err := c.JSONEncode(opts)
	if err != nil {
		return UserResponse{}, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		endpoint,
		bytes.NewBuffer(data),
	)
	if err != nil {
		return UserResponse{}, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("User-Agent", "workos-go/"+workos.Version)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return UserResponse{}, err
	}
	defer res.Body.Close()

	if err = workos_errors.TryGetHTTPError(res); err != nil {
		return UserResponse{}, err
	}

	var body UserResponse
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&body)

	return body, err
}

// CreatePasswordResetChallenge creates a password reset challenge and emails a password reset link to an
// unmanaged user.
func (c *Client) CreatePasswordResetChallenge(ctx context.Context, opts CreatePasswordResetChallengeOpts) (UserResponse, error) {
	endpoint := fmt.Sprintf(
		"%s/users/password_reset_challenge",
		c.Endpoint,
	)

	data, err := c.JSONEncode(opts)
	if err != nil {
		return UserResponse{}, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		endpoint,
		bytes.NewBuffer(data),
	)
	if err != nil {
		return UserResponse{}, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("User-Agent", "workos-go/"+workos.Version)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return UserResponse{}, err
	}
	defer res.Body.Close()

	if err = workos_errors.TryGetHTTPError(res); err != nil {
		return UserResponse{}, err
	}

	var body UserResponse
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&body)

	return body, err
}

// CompletePasswordReset resets user password using token that was sent to the user.
func (c *Client) CompletePasswordReset(ctx context.Context, opts CompletePasswordResetOpts) (User, error) {
	endpoint := fmt.Sprintf(
		"%s/users/password_reset",
		c.Endpoint,
	)

	data, err := c.JSONEncode(opts)
	if err != nil {
		return User{}, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		endpoint,
		bytes.NewBuffer(data),
	)
	if err != nil {
		return User{}, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("User-Agent", "workos-go/"+workos.Version)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return User{}, err
	}
	defer res.Body.Close()

	if err = workos_errors.TryGetHTTPError(res); err != nil {
		return User{}, err
	}

	var body User
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&body)

	return body, err
}

// SendMagicAuthCode creates a one-time Magic Auth code and emails it to the user.
func (c *Client) SendMagicAuthCode(ctx context.Context, opts SendMagicAuthCodeOpts) (User, error) {
	endpoint := fmt.Sprintf(
		"%s/users/magic_auth/send",
		c.Endpoint,
	)

	data, err := c.JSONEncode(opts)
	if err != nil {
		return User{}, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		endpoint,
		bytes.NewBuffer(data),
	)
	if err != nil {
		return User{}, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("User-Agent", "workos-go/"+workos.Version)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return User{}, err
	}
	defer res.Body.Close()

	if err = workos_errors.TryGetHTTPError(res); err != nil {
		return User{}, err
	}

	var body User
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&body)

	return body, err
}