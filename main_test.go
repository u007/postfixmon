package main

import (
	"fmt"
	"testing"
)

// Initialize the log function for tests.
// This is necessary because emailDomainName calls the global 'log' function.
func init() {
	log = func(msg string, args ...interface{}) {
		// For tests, we can simply print to console or do nothing.
		// Using fmt.Printf to mimic the behavior of the original log function.
		fmt.Printf("test-log: "+msg+"\n", args...)
	}
}

func TestEmailDomainName(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		want    string
		wantErr bool
	}{
		{
			name:    "valid email",
			email:   "user@example.com",
			want:    "example.com",
			wantErr: false,
		},
		{
			name:    "email with subdomain",
			email:   "user@sub.example.co.uk",
			want:    "sub.example.co.uk",
			wantErr: false,
		},
		{
			name:    "invalid email - no @",
			email:   "userexample.com",
			want:    "",
			wantErr: true,
		},
		{
			name:    "email with leading/trailing spaces",
			email:   "  user@example.com  ",
			want:    "example.com",
			wantErr: false,
		},
		{
			name:    "email with leading spaces only",
			email:   " user@example.com",
			want:    "example.com",
			wantErr: false,
		},
		{
			name:    "email with trailing spaces only",
			email:   "user@example.com ",
			want:    "example.com",
			wantErr: false,
		},
		{
			name:    "empty email string",
			email:   "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "email string without @ but with content",
			email:   "user",
			want:    "",
			wantErr: true,
		},
		{
			name:    "email string with @ at the beginning",
			email:   "@example.com",
			want:    "example.com",
			wantErr: false, // Assuming an empty local part is valid for domain extraction
		},
		{
			name:    "email string with @ at the end",
			email:   "user@",
			want:    "",
			wantErr: true, // Domain part is empty
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := emailDomainName(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("emailDomainName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("emailDomainName() = %v, want %v", got, tt.want)
			}
		})
	}
}
