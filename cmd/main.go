package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/cheeyeo/AWS_S3_PIPES/pipes"
	"golang.org/x/sync/errgroup"
)

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

// Download stream
func Download(pipe string, bucket string, key string, file string) {
	g, ctx := errgroup.WithContext(context.Background())
	pipeInput := &pipes.DownloadInput{}
	pipeOutput := &pipes.DownloadOutput{File: file}

	// Creates 2 go routines
	// One to read from S3 to source end of the pipe
	// One to read from target end of pipe either to local file
	// if file arg specified; else it loops waiting to be read via STDIN
	g.Go(func() error {
		c := pipes.NewPipeWithCancellation(pipeOutput)
		return c.Stream(ctx, pipe, bucket, key)
	})

	g.Go(func() error {
		c := pipes.NewPipeWithCancellation(pipeInput)
		return c.Stream(ctx, pipe, bucket, key)
	})

	if err := g.Wait(); err != nil {
		exitErrorf("Error in download: %v\n", err)
	}

	fmt.Printf("Downloaded File: %d bytes\n", pipeInput.FSize)
}

func Upload(pipe string, bucket string, key string, file string) {
	g, ctx := errgroup.WithContext(context.Background())

	uploadInput := &pipes.UploadInput{
		UploadFile: file,
	}
	uploadOutput := &pipes.UploadOutput{}

	// Creates 2 go routines
	// 1. Reads from local file and copies to pipe
	// 2. Reads from target end of pipe and uploads to S3
	g.Go(func() error {
		c := pipes.NewPipeWithCancellation(uploadInput)
		return c.Stream(ctx, pipe, bucket, key)
	})

	g.Go(func() error {
		c := pipes.NewPipeWithCancellation(uploadOutput)
		return c.Stream(ctx, pipe, bucket, key)
	})

	if err := g.Wait(); err != nil {
		exitErrorf("Error in upload: %v\n", err)
	}

	fmt.Printf("Upload Successful\n\n")
	fmt.Printf("Location: %s\nVersionID: %s\nUploadID: %s\nETag: %s\n", uploadOutput.Location, uploadOutput.VersionID, uploadOutput.UploadID, uploadOutput.ETag)
}

func main() {
	downloadCmd := flag.NewFlagSet("download", flag.ExitOnError)
	downloadBucket := downloadCmd.String("bucket", "", "S3 Bucket to download from...")
	downloadKey := downloadCmd.String("key", "", "S3 Key to download from...")
	downloadPipe := downloadCmd.String("pipe", "", "Named pipe to download to...")
	downloadFile := downloadCmd.String("file", "", "Filename to save the stream output.")

	uploadCmd := flag.NewFlagSet("upload", flag.ExitOnError)
	uploadFile := uploadCmd.String("file", "", "Local file to upload")
	uploadBucket := uploadCmd.String("bucket", "", "S3 bucket to upload to...")
	uploadKey := uploadCmd.String("key", "", "S3 file name of uploaded file")
	uploadPipe := uploadCmd.String("pipe", "", "Named pipe to upload to...")

	if len(os.Args) < 2 {
		fmt.Println("Expect 'upload' or 'download' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "download":
		downloadCmd.Parse(os.Args[2:])
		fmt.Printf("Bucket: %s\n", *downloadBucket)
		fmt.Printf("Key: %s\n", *downloadKey)
		fmt.Printf("Pipe: %s\n", *downloadPipe)
		fmt.Println()

		Download(*downloadPipe, *downloadBucket, *downloadKey, *downloadFile)
	case "upload":
		uploadCmd.Parse(os.Args[2:])
		fmt.Printf("Bucket: %s\n", *uploadBucket)
		fmt.Printf("Key: %s\n", *uploadKey)
		fmt.Printf("Pipe: %s\n", *uploadPipe)
		fmt.Printf("Local file: %s\n\n", *uploadFile)

		Upload(*uploadPipe, *uploadBucket, *uploadKey, *uploadFile)
	default:
		fmt.Println("Subcommand of either 'download' or 'upload' is required")
		os.Exit(1)
	}
}
