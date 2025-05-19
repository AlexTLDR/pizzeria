package auth

import (
	"testing"

	"golang.org/x/oauth2"
)

func TestOAuthConfig_IsAllowedEmail(t *testing.T) {
	tests := []struct {
		name          string
		allowedEmails []string
		email         string
		want          bool
	}{
		{
			name:          "Email in allowed list",
			allowedEmails: []string{"test@example.com", "admin@example.com", "owner@example.com"},
			email:         "admin@example.com",
			want:          true,
		},
		{
			name:          "Email not in allowed list",
			allowedEmails: []string{"test@example.com", "admin@example.com", "owner@example.com"},
			email:         "unauthorized@example.com",
			want:          false,
		},
		{
			name:          "Empty email",
			allowedEmails: []string{"test@example.com", "admin@example.com"},
			email:         "",
			want:          false,
		},
		{
			name:          "Case sensitive comparison",
			allowedEmails: []string{"Test@Example.com"},
			email:         "test@example.com",
			want:          false,
		},
		{
			name:          "Single allowed email",
			allowedEmails: []string{"single@example.com"},
			email:         "single@example.com",
			want:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			c := &OAuthConfig{
				GoogleOAuthConfig: &oauth2.Config{},
				AllowedEmails:     tt.allowedEmails,
			}

			if got := c.IsAllowedEmail(tt.email); got != tt.want {
				t.Errorf("OAuthConfig.IsAllowedEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}