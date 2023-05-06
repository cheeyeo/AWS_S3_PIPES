// Data structure for download

package pipes

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/cheeyeo/AWS_S3_PIPES/files"
	"github.com/cheeyeo/AWS_S3_PIPES/s3helpers"
	"github.com/cheeyeo/AWS_S3_PIPES/writer"
	"github.com/schollz/progressbar/v3"
)

// Represents input for the download
type DownloadInput struct {
	FSize int64
}

// Represents target for the download
// Can be to a local file or nil
type DownloadOutput struct {
	File string
}

func (pi *DownloadInput) Fetch(ctx context.Context, pipe string, bucket string, key string) error {

	// Check bucket exists and we can access it
	exists, err := s3helpers.BucketValidator(bucket)
	if !exists {
		return fmt.Errorf("DownloadInput: Unable to locate bucket: %v\n", err)
	}

	size, err := files.GetS3FileSize(bucket, key)
	if err != nil {
		return fmt.Errorf("DownloadInput: Unable to parse S3 file: %v\n", err)
	}

	pipeFile, err := os.OpenFile(pipe, os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("DownloadInput: Unable to read pipe file: %v\n", err)
	}
	defer pipeFile.Close()

	fSize, err := writer.PipeDownload(ctx, bucket, key, pipeFile, size)
	if err != nil {
		return fmt.Errorf("DownloadInput: Unable to run pipe download: %v\n", err)
	}
	pi.FSize = fSize

	return nil
}

func (pi *DownloadOutput) Fetch(ctx context.Context, pipe string, bucket string, key string) error {
	if len(pi.File) > 0 {
		savedFile, err := os.Create(pi.File)
		if err != nil {
			return fmt.Errorf("DownloadOutput: Error creating file: %v\n", err)
		}
		defer savedFile.Close()

		source, err := os.OpenFile(pipe, os.O_RDONLY, 0640)
		if err != nil {
			return fmt.Errorf("DownloadOutput: Error opening named pipe: %v\n", err)
		}
		defer source.Close()

		size, err := files.GetS3FileSize(bucket, key)
		if err != nil {
			return fmt.Errorf("DownloadOutput: Unable to parse S3 file: %v\n", err)
		}

		// Creates download progressbar...
		downloadMsg := fmt.Sprintf("Downloading %s", key)
		bar := progressbar.DefaultBytes(
			size,
			downloadMsg,
		)

		_, err = io.Copy(io.MultiWriter(savedFile, bar), source)
		if err != nil {
			return fmt.Errorf("DownloadOutput: Error downloading from named pipe: %v\n", err)
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
