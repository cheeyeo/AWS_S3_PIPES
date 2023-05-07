package tests

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/cheeyeo/AWS_S3_PIPES/writer"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/awstesting"
	"github.com/aws/aws-sdk-go/awstesting/unit"
)

func TestWriterPipeDownload_ContextCancelled(t *testing.T) {
	var m sync.Mutex
	sess := unit.Session
	sess.Handlers.Send.Clear()
	sess.Handlers.Send.PushBack(func(r *request.Request) {
		m.Lock()
		defer m.Unlock()

		r.HTTPResponse = &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte{})),
			Header:     http.Header{},
		}
		r.HTTPResponse.Header.Set("Content-Length", fmt.Sprintf("%d", 1))
	})

	ctx := &awstesting.FakeContext{DoneCh: make(chan struct{})}
	ctx.Error = fmt.Errorf("context canceled")
	close(ctx.DoneCh)

	pipe, _, _ := CreateTempFile(t, 1)

	_, err := writer.PipeDownload(ctx, sess, "Bucket", "Key", pipe, 1)
	if err == nil {
		t.Errorf("Expected error but got nil")
	}

	aerr := err.(awserr.Error)
	if e, a := "RequestCanceled", aerr.Code(); e != a {
		t.Errorf("expected error code %q, got %q", e, a)
	}
	if e, a := "request context canceled", aerr.Message(); !strings.Contains(a, e) {
		t.Errorf("expected error message to contain %q, but did not %q", e, a)
	}
}

func TestWriterPipeDownload(t *testing.T) {
	var data = make([]byte, 1024*1024*2)
	var m sync.Mutex
	// Below clears the session handlers and prevent S3 from making actual s3 calls
	sess := unit.Session
	sess.Handlers.Send.Clear()
	sess.Handlers.Send.PushBack(func(r *request.Request) {
		m.Lock()
		defer m.Unlock()

		rerng := regexp.MustCompile(`bytes=(\d+)-(\d+)`)
		rng := rerng.FindStringSubmatch(r.HTTPRequest.Header.Get("Range"))

		start, _ := strconv.ParseInt(rng[1], 10, 64)
		fin, _ := strconv.ParseInt(rng[2], 10, 64)
		fin++

		if fin > int64(len(data)) {
			fin = int64(len(data))
		}

		bodyBytes := data[start:fin]

		r.HTTPResponse = &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewReader(bodyBytes)),
			Header:     http.Header{},
		}
		r.HTTPResponse.Header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, fin-1, len(data)))
		r.HTTPResponse.Header.Set("Content-Length", fmt.Sprintf("%d", len(bodyBytes)))
	})

	pipe, _, _ := CreateTempFile(t, 1)

	n, err := writer.PipeDownload(context.Background(), sess, "Bucket", "Key", pipe, 1)

	if err != nil {
		t.Errorf("Expected error but got nil")
	}

	if n != int64(len(data)) {
		t.Errorf("Expected %d bytes, Got %d", len(data), n)
	}
}
