### AWS S3 Pipe

Utility to perform S3 uploads and downloads using named pipes with the AWS GO SDK

Supports:
* Stream upload via named pipes

* Stream download via named pipes


### Run

Run `make build` which installs the binary into `/tmp/bin/s3pipe`

You need to ensure that the named pipe exists first; if not run:
```
mkfifo <NAME OF PIPE>
```

Also ensure your AWS credentials are reachable via:
```
export AWS_PROFILE=XXX

export AWS_REGION=XXX
```


It supports only 2 actions `upload` and `download` with the same parameters:
```
bucket - Name of S3 bucket
key - Name of file in S3 bucket
pipe - Name of local named pipe
file - Optional input. If given, it will setup the pipe stream and upload / download the file automatically.
```

For instance, given that we have a named pipe of `pipe1` we can setup for upload as such:
```
go run cmd/main.go upload --bucket BUCKET --key KEY --pipe pipe1 --file myfile
```

If the `file` parameter is omitted, the program will block as its waiting for input on the receiving end of the pipe. You need to open another terminal to read/write from the pipe.

Given the same example as above again but without the `file` parameter:
```
go run cmd/main.go upload --bucket BUCKET --key KEY --pipe pipe1
```

We can publish data onto the pipe in another terminal:
```
cat myfile > pipe1
```

This will trigger the cli to continue with the upload.

There are two client examples in the `examples` folder that you can test with in a separate terminal.


### Develop

You need to install the dependencies first:
```
go mod download
go mod tidy -v
```

The core cli is in `cmd/main.go`

The packages are organized as such:

* files - Contain file helpers such as detecting file sizes 

* http - Custom module to override the HTTP client settings for AWS clients

* pipes - Wrapper code that delegates to either the `reader` or `writer` package to upload/download files

* reader - Contains the actual S3 reader code.

* writer - Actual S3 writer code.

* s3helpers - Helpers for S3 service

The tests are located in the `tests` directory.

To run the tests:
```
make test
```

### Build

```
make build
```

This will build the binary into `/tmp/bin/s3pipe`