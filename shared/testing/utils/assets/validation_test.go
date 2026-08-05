package assets_test

import (
	"errors"
	"testing"

	"github.com/TheAmgadX/moltaqa-backend/shared/utils/assets"
)

func TestValidateProfileImagePath_NilOrEmpty(t *testing.T) {
	if err := assets.ValidateProfileImagePath(""); err != nil {
		t.Fatalf("expected nil for empty path, got %v", err)
	}
}

func TestValidateProfileImagePath_ValidPaths(t *testing.T) {
	validPaths := []string{
		"uploads/avatar.jpg",
		"images/profile.jpeg",
		"avatar.png",
		"user_123.webp",
		"UPLOADS/AVATAR.PNG",
	}

	for _, path := range validPaths {
		t.Run("valid path: "+path, func(t *testing.T) {
			if err := assets.ValidateProfileImagePath(path); err != nil {
				t.Fatalf("expected path %q to be valid, got %v", path, err)
			}
		})
	}
}

func TestValidateProfileImagePath_InvalidPaths(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr error
	}{
		{
			name:    "path too long",
			path:    string(make([]byte, 501)),
			wantErr: assets.ErrInvalidProfileImagePath,
		},
		{
			name:    "path traversal dot-dot",
			path:    "uploads/../secret.txt",
			wantErr: assets.ErrInvalidProfileImagePath,
		},
		{
			name:    "absolute path prefix",
			path:    "/etc/passwd",
			wantErr: assets.ErrInvalidProfileImagePath,
		},
		{
			name:    "unallowed extension gif",
			path:    "avatar.gif",
			wantErr: assets.ErrInvalidProfileImageType,
		},
		{
			name:    "unallowed extension txt",
			path:    "document.txt",
			wantErr: assets.ErrInvalidProfileImageType,
		},
		{
			name:    "no extension",
			path:    "no-extension-file",
			wantErr: assets.ErrInvalidProfileImageType,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotErr := assets.ValidateProfileImagePath(tc.path)
			if !errors.Is(gotErr, tc.wantErr) {
				t.Fatalf("path %q: want error %v, got %v", tc.path, tc.wantErr, gotErr)
			}
		})
	}
}
