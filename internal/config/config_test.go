package config

import (
	"testing"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		c       Config
		wantErr bool
	}{
		{
			name: "given non-empty values, expect no errors",
			c: Config{
				AWSRegion:     "us-west-2",
				AWSBucketName: "testbucket",
			},
		},
		{
			name: "given empty region, expect error",
			c: Config{
				AWSBucketName: "testbucket",
			},
			wantErr: true,
		},
		{
			name: "given empty bucket, expect error",
			c: Config{
				AWSRegion: "us-west-2",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateConfig(tt.c); (err != nil) != tt.wantErr {
				t.Errorf("ValidateOptions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
