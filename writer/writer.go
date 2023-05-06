package writer

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync/atomic"

	"github.com/cheeyeo/AWS_S3_PIPES/files"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type s3WriterAt struct {
	written int64
	w       io.Writer
	size    int64
}

func (p *s3WriterAt) WriteAt(b []byte, off int64) (n int, err error) {
	atomic.AddInt64(&p.written, int64(len(b)))
	// percentageDownloaded := float32(p.written*100) / float32(p.size)
	// fmt.Printf("File size: %d, downloaded: %d, percentage: %.2f%%\r", p.size, p.written, percentageDownloaded)

	return p.w.Write(b)
}

func PipeDownload(ctx context.Context, bucket string, source string, pipeFile *os.File, fileSize int64) (int64, error) {
	// Reads the source from the bucket and stream it into pipeFile
	sess := session.Must(session.NewSession())
	downloader := s3manager.NewDownloader(sess, func(d *s3manager.Downloader) {
		d.Concurrency = 10
		d.PartSize = 25 * 1024 * 1024 // 20MB part size
		d.BufferProvider = s3manager.NewPooledBufferedWriterReadFromProvider(25 * 1024 * 1024)
	})

	fmt.Println("\nStarting Download, Size: ", files.ByteCountDecimal(fileSize))

	writer := &s3WriterAt{w: pipeFile, size: fileSize, written: 0}
	n, err := downloader.DownloadWithContext(ctx, writer, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(source),
	})

	if err != nil {
		return n, err
	}

	return n, nil
}
