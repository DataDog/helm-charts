package utils

import "testing"

func TestImageRepository(t *testing.T) {
	tests := []struct {
		name  string
		image string
		want  string
	}{
		{
			name:  "repository with tag",
			image: "registry.datadoghq.com/operator:1.30.0-rc.2",
			want:  "registry.datadoghq.com/operator",
		},
		{
			name:  "repository without tag",
			image: "registry.datadoghq.com/operator",
			want:  "registry.datadoghq.com/operator",
		},
		{
			name:  "repository with tag and digest",
			image: "registry.datadoghq.com/operator:1.18.0@sha256:0000",
			want:  "registry.datadoghq.com/operator",
		},
		{
			name:  "registry host with a port",
			image: "myregistry:5000/operator:1.2.3",
			want:  "myregistry:5000/operator",
		},
		{
			name:  "registry host with a port, no tag",
			image: "myregistry:5000/operator",
			want:  "myregistry:5000/operator",
		},
		{
			name:  "no registry host, just repo and tag",
			image: "operator:1.2.3",
			want:  "operator",
		},
		{
			name:  "repository with digest, no tag",
			image: "registry.datadoghq.com/operator@sha256:0000",
			want:  "registry.datadoghq.com/operator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ImageRepository(tt.image); got != tt.want {
				t.Errorf("ImageRepository(%q) = %q, want %q", tt.image, got, tt.want)
			}
		})
	}
}
