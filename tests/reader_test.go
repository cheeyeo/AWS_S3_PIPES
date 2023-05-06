package tests

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/cheeyeo/AWS_S3_PIPES/reader"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/awstesting"
	"github.com/aws/aws-sdk-go/awstesting/unit"
)

const respMsg = `<?xml version="1.0" encoding="UTF-8"?>
<CompleteMultipartUploadOutput>
   <Location>mockValue</Location>
   <Bucket>mockValue</Bucket>
   <Key>mockValue</Key>
   <ETag>mockValue</ETag>
</CompleteMultipartUploadOutput>`

func createTempFile(t *testing.T, size int64) (*os.File, func(*testing.T), error) {
	file, err := ioutil.TempFile(os.TempDir(), aws.SDKName+t.Name())
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

func TestReaderPipeUpload_Fail(t *testing.T) {
	sess := unit.Session

	ctx := &awstesting.FakeContext{DoneCh: make(chan struct{})}
	ctx.Error = fmt.Errorf("context canceled")
	close(ctx.DoneCh)

	pipe, _, _ := createTempFile(t, 1)

	_, err := reader.PipeUpload(ctx, sess, "Bucket", "Key", pipe)
	if err == nil {
		t.Errorf("Expected error but got nil")
	}

	aerr := err.(awserr.Error)
	if e, a := "ReadRequestBody", aerr.Code(); e != a {
		t.Errorf("expected error code %q, got %q", e, a)
	}
	if e, a := "read upload data failed", aerr.Message(); !strings.Contains(a, e) {
		t.Errorf("expected error message to contain %q, but did not %q", e, a)
	}
}

func TestReaderPipeUpload_Pass(t *testing.T) {
	// Below clears the session handlers and prevent S3 from making actual s3 calls
	sess := unit.Session
	sess.Handlers.Unmarshal.Clear()
	sess.Handlers.UnmarshalMeta.Clear()
	sess.Handlers.UnmarshalError.Clear()
	sess.Handlers.ValidateResponse.Clear()
	sess.Handlers.Send.Clear()

	sess.Handlers.Send.PushBack(func(r *request.Request) {
		r.HTTPResponse = &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(respMsg))),
		}
	})

	pipe, _, _ := createTempFile(t, 1)

	_, err := reader.PipeUpload(context.Background(), sess, "Bucket", "Key", pipe)

	if err != nil {
		t.Errorf("Expected error but got nil")
	}
}
