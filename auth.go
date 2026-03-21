package qai

import "context"

// AuthUser contains user information returned after authentication.
type AuthUser struct {
	// ID is the user identifier.
	ID string `json:"id"`

	// Name is the display name.
	Name string `json:"name,omitempty"`

	// Email is the email address.
	Email string `json:"email,omitempty"`

	// AvatarURL is the avatar image URL.
	AvatarURL string `json:"avatar_url,omitempty"`
}

// AuthResponse is the response from authentication endpoints.
type AuthResponse struct {
	// Token is the API token for subsequent requests.
	Token string `json:"token"`

	// User is the authenticated user information.
	User AuthUser `json:"user"`
}

// AuthAppleRequest is the request body for Apple Sign-In.
type AuthAppleRequest struct {
	// IDToken is the Apple identity token (JWT from Sign in with Apple).
	IDToken string `json:"id_token"`

	// Name is the optional display name (only provided on first sign-in).
	Name string `json:"name,omitempty"`
}

// AuthApple authenticates with Apple Sign-In.
// The IDToken is the JWT received from the Sign in with Apple flow.
// On first sign-in, pass the user's Name so the account is created with a display name.
func (c *Client) AuthApple(ctx context.Context, req *AuthAppleRequest) (*AuthResponse, error) {
	var resp AuthResponse
	_, err := c.doJSON(ctx, "POST", "/qai/v1/auth/apple", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
