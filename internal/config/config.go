package config

import (
	"errors"

	"github.com/hashicorp/go-multierror"
)

type Config struct {
	IOStreams     IOStreams
	AWSRegion     string `env:"AWS_REGION" envDefault:"us-west-2"`
	AWSBucketName string `env:"AWS_BUCKET" envDefault:"github-ghostsquad-s3-file-explorer"`
	Port          int    `env:"PORT" envDefault:"8080"`
	BindAddress   string `env:"BIND_ADDRESS"`
}

func Empty(val string) bool {
	return val == ""
}

func ValidateConfig(g Config) error {
	mErr := &multierror.Error{}

	if Empty(g.AWSRegion) {
		mErr = multierror.Append(mErr, errors.New("AWS Region is empty"))
	}

	if Empty(g.AWSBucketName) {
		mErr = multierror.Append(mErr, errors.New("AWS Bucket Name is empty"))
	}

	return mErr.ErrorOrNil()
}
