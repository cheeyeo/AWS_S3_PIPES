package tests

import (
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
)

func CreateTempFile(t *testing.T, size int64) (*os.File, func(*testing.T), error) {
	file, err := os.CreateTemp(os.TempDir(), aws.SDKName+t.Name())
	if err != nil {
		return nil, nil, err
	}
	filename := file.Name()
	if err := file.Truncate(size); err != nil {
		return nil, nil, err
	}

	return file,
		func(t *testing.T) {
			if err := file.Close(); err != nil {
				t.Errorf("failed to close temp file, %s, %v", filename, err)
			}
			if err := os.Remove(filename); err != nil {
				t.Errorf("failed to remove temp file, %s, %v", filename, err)
			}
		},
		nil
}
