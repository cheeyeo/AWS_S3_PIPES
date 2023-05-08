// Data structure for download

package pipes

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/cheeyeo/AWS_S3_PIPES/s3helpers"
	"github.com/cheeyeo/AWS_S3_PIPES/writer"
	"github.com/schollz/progressbar/v3"
)

// DownloadInput represents source of pipe
type DownloadInput struct {
	FSize int64
}

// DownloadOutput represents target for the download
type DownloadOutput struct {
	File string
}

func (pi *DownloadInput) Stream(ctx context.Context, pipe string, bucket string, key string) error {
	sess := session.Must(session.NewSession())

	// Check bucket exists and we can access it
	exists, err := s3helpers.BucketValidator(sess, bucket)
	if !exists {
		return fmt.Errorf("DownloadInput: Unable to locate bucket: %v", err)
	}

	pipeFile, err := os.OpenFile(pipe, os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("DownloadInput: Unable to read pipe file: %v", err)
	}
	defer pipeFile.Close()

	fileSize, err := s3helpers.GetS3FileSize(sess, bucket, key)
	if err != nil {
		return fmt.Errorf("DownloadOutput: Unable to parse S3 file: %v", err)
	}
	fSize, err := writer.PipeDownload(ctx, sess, bucket, key, pipeFile, fileSize)
	if err != nil {
		return fmt.Errorf("DownloadInput: Unable to run pipe download: %v", err)
	}
	pi.FSize = fSize

	return nil
}

func (pi *DownloadOutput) Stream(ctx context.Context, pipe string, bucket string, key string) error {
	if len(pi.File) > 0 {
		savedFile, err := os.Create(pi.File)
		if err != nil {
			return fmt.Errorf("DownloadOutput: Error creating file: %v", err)
		}
		defer savedFile.Close()

		source, err := os.OpenFile(pipe, os.O_RDONLY, 0640)
		if err != nil {
			return fmt.Errorf("DownloadOutput: Error opening named pipe: %v", err)
		}
		defer source.Close()

		sess := session.Must(session.NewSession())
		size, err := s3helpers.GetS3FileSize(sess, bucket, key)
		if err != nil {
			return fmt.Errorf("DownloadOutput: Unable to parse S3 file: %v", err)
		}

		// Creates download progressbar...
		downloadMsg := fmt.Sprintf("Downloading %s", key)
		bar := progressbar.DefaultBytes(
			size,
			downloadMsg,
		)

		_, err = io.Copy(io.MultiWriter(savedFile, bar), source)
		if err != nil {
			return fmt.Errorf("DownloadOutput: Error downloading from named pipe: %v", err)
		}
	} else {
		instruct := `
		No file download target has been specified.
		Read from the pipe manually like so:
	
		cat %s  > myfile.txt
		`
		fmt.Printf(instruct, pipe)
		fmt.Println()
	}

	return nil
}
