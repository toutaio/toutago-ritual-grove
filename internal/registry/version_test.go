package registry

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    Version
		wantErr bool
	}{
		{
			name:    "valid semver",
			version: "1.2.3",
			want:    Version{Major: 1, Minor: 2, Patch: 3},
			wantErr: false,
		},
		{
			name:    "valid semver with v prefix",
			version: "v2.0.0",
			want:    Version{Major: 2, Minor: 0, Patch: 0},
			wantErr: false,
		},
		{
			name:    "zero version",
			version: "0.0.0",
			want:    Version{Major: 0, Minor: 0, Patch: 0},
			wantErr: false,
		},
		{
			name:    "large version numbers",
			version: "999.888.777",
			want:    Version{Major: 999, Minor: 888, Patch: 777},
			wantErr: false,
		},
		{
			name:    "invalid - missing patch",
			version: "1.2",
			wantErr: true,
		},
		{
			name:    "invalid - non-numeric",
			version: "1.2.x",
			wantErr: true,
		},
		{
			name:    "invalid - negative number",
			version: "1.-2.3",
			wantErr: true,
		},
		{
			name:    "invalid - empty string",
			version: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVersion(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsVersionNewer(t *testing.T) {
	tests := []struct {
		name           string
		newVersion     string
		currentVersion string
		want           bool
		wantErr        bool
	}{
		{
			name:           "newer major version",
			newVersion:     "2.0.0",
			currentVersion: "1.0.0",
			want:           true,
			wantErr:        false,
		},
		{
			name:           "newer minor version",
			newVersion:     "1.2.0",
			currentVersion: "1.1.0",
			want:           true,
			wantErr:        false,
		},
		{
			name:           "newer patch version",
			newVersion:     "1.0.2",
			currentVersion: "1.0.1",
			want:           true,
			wantErr:        false,
		},
		{
			name:           "same version",
			newVersion:     "1.0.0",
			currentVersion: "1.0.0",
			want:           false,
			wantErr:        false,
		},
		{
			name:           "older major version",
			newVersion:     "1.0.0",
			currentVersion: "2.0.0",
			want:           false,
			wantErr:        false,
		},
		{
			name:           "older minor version",
			newVersion:     "1.1.0",
			currentVersion: "1.2.0",
			want:           false,
			wantErr:        false,
		},
		{
			name:           "older patch version",
			newVersion:     "1.0.1",
			currentVersion: "1.0.2",
			want:           false,
			wantErr:        false,
		},
		{
			name:           "invalid new version",
			newVersion:     "invalid",
			currentVersion: "1.0.0",
			want:           false,
			wantErr:        true,
		},
		{
			name:           "invalid current version",
			newVersion:     "1.0.0",
			currentVersion: "invalid",
			want:           false,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsVersionNewer(tt.newVersion, tt.currentVersion)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsVersionNewer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsVersionNewer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersionString(t *testing.T) {
	tests := []struct {
		name    string
		version Version
		want    string
	}{
		{
			name:    "standard version",
			version: Version{Major: 1, Minor: 2, Patch: 3},
			want:    "1.2.3",
		},
		{
			name:    "zero version",
			version: Version{Major: 0, Minor: 0, Patch: 0},
			want:    "0.0.0",
		},
		{
			name:    "large numbers",
			version: Version{Major: 999, Minor: 888, Patch: 777},
			want:    "999.888.777",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.version.String(); got != tt.want {
				t.Errorf("Version.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersionCompare(t *testing.T) {
	tests := []struct {
		name string
		v1   Version
		v2   Version
		want int
	}{
		{
			name: "equal versions",
			v1:   Version{1, 2, 3},
			v2:   Version{1, 2, 3},
			want: 0,
		},
		{
			name: "v1 greater major",
			v1:   Version{2, 0, 0},
			v2:   Version{1, 9, 9},
			want: 1,
		},
		{
			name: "v1 lesser major",
			v1:   Version{1, 9, 9},
			v2:   Version{2, 0, 0},
			want: -1,
		},
		{
			name: "v1 greater minor",
			v1:   Version{1, 2, 0},
			v2:   Version{1, 1, 9},
			want: 1,
		},
		{
			name: "v1 lesser minor",
			v1:   Version{1, 1, 9},
			v2:   Version{1, 2, 0},
			want: -1,
		},
		{
			name: "v1 greater patch",
			v1:   Version{1, 2, 4},
			v2:   Version{1, 2, 3},
			want: 1,
		},
		{
			name: "v1 lesser patch",
			v1:   Version{1, 2, 3},
			v2:   Version{1, 2, 4},
			want: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v1.Compare(tt.v2); got != tt.want {
				t.Errorf("Version.Compare() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersionConstraintSatisfies(t *testing.T) {
	tests := []struct {
		name       string
		constraint string
		version    string
		want       bool
		wantErr    bool
	}{
		{
			name:       "exact match",
			constraint: "=1.2.3",
			version:    "1.2.3",
			want:       true,
		},
		{
			name:       "exact no match",
			constraint: "=1.2.3",
			version:    "1.2.4",
			want:       false,
		},
		{
			name:       "greater than satisfied",
			constraint: ">1.0.0",
			version:    "1.0.1",
			want:       true,
		},
		{
			name:       "greater than not satisfied",
			constraint: ">1.0.0",
			version:    "1.0.0",
			want:       false,
		},
		{
			name:       "greater or equal satisfied (greater)",
			constraint: ">=1.0.0",
			version:    "1.0.1",
			want:       true,
		},
		{
			name:       "greater or equal satisfied (equal)",
			constraint: ">=1.0.0",
			version:    "1.0.0",
			want:       true,
		},
		{
			name:       "greater or equal not satisfied",
			constraint: ">=1.0.0",
			version:    "0.9.9",
			want:       false,
		},
		{
			name:       "less than satisfied",
			constraint: "<2.0.0",
			version:    "1.9.9",
			want:       true,
		},
		{
			name:       "less than not satisfied",
			constraint: "<2.0.0",
			version:    "2.0.0",
			want:       false,
		},
		{
			name:       "less or equal satisfied (less)",
			constraint: "<=2.0.0",
			version:    "1.9.9",
			want:       true,
		},
		{
			name:       "less or equal satisfied (equal)",
			constraint: "<=2.0.0",
			version:    "2.0.0",
			want:       true,
		},
		{
			name:       "less or equal not satisfied",
			constraint: "<=2.0.0",
			version:    "2.0.1",
			want:       false,
		},
		{
			name:       "invalid constraint operator",
			constraint: "~1.0.0",
			version:    "1.0.1",
			wantErr:    true,
		},
		{
			name:       "invalid constraint format",
			constraint: ">=",
			version:    "1.0.0",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := ParseConstraint(tt.constraint)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseConstraint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			version, err := ParseVersion(tt.version)
			if err != nil {
				t.Fatalf("ParseVersion() failed: %v", err)
			}

			if got := constraint.Satisfies(version); got != tt.want {
				t.Errorf("Constraint.Satisfies() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSelectBestVersion(t *testing.T) {
	tests := []struct {
		name       string
		versions   []string
		constraint string
		want       string
		wantErr    bool
	}{
		{
			name:       "select highest matching",
			versions:   []string{"1.0.0", "1.1.0", "1.2.0", "2.0.0"},
			constraint: "<2.0.0",
			want:       "1.2.0",
		},
		{
			name:       "select exact",
			versions:   []string{"1.0.0", "1.1.0", "1.2.0"},
			constraint: "=1.1.0",
			want:       "1.1.0",
		},
		{
			name:       "no matching versions",
			versions:   []string{"1.0.0", "1.1.0"},
			constraint: ">=2.0.0",
			wantErr:    true,
		},
		{
			name:       "select from unordered list",
			versions:   []string{"2.0.0", "1.0.0", "1.5.0", "1.2.0"},
			constraint: ">=1.2.0",
			want:       "2.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := ParseConstraint(tt.constraint)
			if err != nil {
				t.Fatalf("ParseConstraint() failed: %v", err)
			}

			got, err := SelectBestVersion(tt.versions, constraint)
			if (err != nil) != tt.wantErr {
				t.Errorf("SelectBestVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("SelectBestVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
