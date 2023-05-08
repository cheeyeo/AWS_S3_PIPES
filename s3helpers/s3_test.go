package s3helpers

import (
	"bytes"

	// "fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/awstesting/unit"
	"github.com/stretchr/testify/assert"
)

func TestBucketValidator(t *testing.T) {
	sess := unit.Session
	sess.Handlers.Send.Clear()

	sess.Handlers.Send.PushBack(func(r *request.Request) {
		r.HTTPResponse = &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte{})),
		}
	})

	res, err := BucketValidator(sess, "Bucket")
	if res != true {
		t.Errorf("Should be %v but received %v", true, res)
	}
	if err != nil {
		t.Errorf("Should not receive error but got error: %v", err)
	}
}

func TestBucketValidatorNotFound(t *testing.T) {
	sess := unit.Session
	sess.Handlers.Send.Clear()

	sess.Handlers.Send.PushBack(func(r *request.Request) {
		r.HTTPResponse = &http.Response{
			StatusCode: 404,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte{})),
		}
	})

	res, err := BucketValidator(sess, "NotFound")
	assert.Equal(t, "Bucket NotFound does not exist", err.Error())
	assert.Equal(t, false, res)

	if res != false {
		t.Errorf("Should be %v but received %v", true, res)
	}
	if err == nil {
		t.Errorf("Should not receive error but got error: %v", err)
	}
}

func TestBucketValidatorForbidden(t *testing.T) {
	sess := unit.Session
	sess.Handlers.Send.Clear()

	sess.Handlers.Send.PushBack(func(r *request.Request) {
		r.HTTPResponse = &http.Response{
			StatusCode: 403,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte{})),
		}
	})

	res, err := BucketValidator(sess, "MyBucket")
	assert.Equal(t, "Bucket MyBucket is forbidden", err.Error())
	assert.Equal(t, false, res)

	if res != false {
		t.Errorf("Should be %v but received %v", true, res)
	}
	if err == nil {
		t.Errorf("Should not receive error but got error: %v", err)
	}
}

func TestBucketValidatorBadRequest(t *testing.T) {
	sess := unit.Session
	sess.Handlers.Send.Clear()

	sess.Handlers.Send.PushBack(func(r *request.Request) {
		r.HTTPResponse = &http.Response{
			StatusCode: 400,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte{})),
		}
	})

	res, err := BucketValidator(sess, "MyBucket")
	assert.Equal(t, "Bucket MyBucket is invalid", err.Error())
	assert.Equal(t, false, res)

	if res != false {
		t.Errorf("Should be %v but received %v", true, res)
	}
	if err == nil {
		t.Errorf("Should not receive error but got error: %v", err)
	}
}

func TestGetS3FileSize(t *testing.T) {
	sess := unit.Session
	sess.Handlers.Send.Clear()

	// NOTE: HEAD Request has no response body and the results can be stubbed in the Response Header field
	sess.Handlers.Send.PushBack(func(r *request.Request) {
		resp := map[string][]string{
			"Content-Length": {"12"},
		}
		r.HTTPRequest.Method = "HEAD"
		r.HTTPResponse = &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte{})),
			Header:     resp,
		}
	})

	res, err := GetS3FileSize(sess, "Bucket", "Key")
	assert.Equal(t, int64(12), res)
	assert.Equal(t, nil, err)
}

func TestGetS3FileSizeError(t *testing.T) {
	sess := unit.Session
	sess.Handlers.Send.Clear()

	// NOTE: HEAD Request has no response body and the results can be stubbed in the Response Header field
	sess.Handlers.Send.PushBack(func(r *request.Request) {
		resp := map[string][]string{
			"Content-Length": {"12"},
		}
		r.HTTPRequest.Method = "HEAD"
		r.HTTPResponse = &http.Response{
			StatusCode: 500,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte{})),
			Header:     resp,
		}
	})

	res, err := GetS3FileSize(sess, "Bucket", "Key")
	assert.Equal(t, int64(0), res)
	assert.Equal(t, "Internal Server Error", err.Error())
}
