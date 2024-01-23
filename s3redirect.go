package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var (
	svc     *s3.S3
	bucket  string
	region  string
	address string
)

func presignedURL(bucket string, key string) (string, error) {
	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	return req.Presign(15 * time.Minute)
}

func redirectHandler(bucket string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url, err := presignedURL(bucket, r.URL.Path[1:])
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func run() error {
	flag.StringVar(&bucket, "bucket", "", "S3 bucket name")
	flag.StringVar(&region, "region", "eu-central-1", "AWS region")
	flag.StringVar(&address, "address", ":8080", "Address to listen on")
	flag.Parse()

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	if err != nil {
		return err
	}

	svc = s3.New(sess)
	if bucket == "" {
		flag.Usage()
		os.Exit(1)
	}

	http.HandleFunc("/", redirectHandler(bucket))
	fmt.Fprintf(os.Stderr, "Starting server on %s, serving %s\n", address, bucket)
	return http.ListenAndServe(address, nil)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
