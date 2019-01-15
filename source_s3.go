package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const ImageSourceTypeS3 ImageSourceType = "s3"

type S3ImageSource struct {
	Config *SourceConfig
}

func NewS3ImageSource(config *SourceConfig) ImageSource {
	return &S3ImageSource{config}
}

func (s *S3ImageSource) Matches(r *http.Request) bool {
	urlQuery := r.URL.Query().Get(URLQueryKey)
	return r.Method == http.MethodGet && strings.HasPrefix(urlQuery, "s3://")
}

func (s *S3ImageSource) GetImage(req *http.Request) ([]byte, error) {
	url, err := parseURL(req)
	if err != nil {
		return nil, ErrInvalidImageURL
	}

	if shouldRestrictOrigin(url, s.Config.AllowedOrigins) {
		return nil, fmt.Errorf("not allowed remote URL origin: %s", url.Host)
	}

	return s.fetchImage(url, req)

}

func (s *S3ImageSource) fetchImage(url *url.URL, ireq *http.Request) ([]byte, error) {
	sess := session.New()
	svc := s3.New(sess, aws.NewConfig())

	// Check remote image size by fetching object size
	if s.Config.MaxAllowedSize > 0 {
		res, err := svc.HeadObject(&s3.HeadObjectInput{
			Bucket: aws.String(url.Host),
			Key:    aws.String(url.Path),
		})
		if err != nil {
			return nil, fmt.Errorf("error fetching image headers from S3: %v", err)
		}

		if int(*res.ContentLength) > s.Config.MaxAllowedSize {
			return nil, fmt.Errorf("Content-Length %d exceeds maximum allowed %d bytes", *res.ContentLength, s.Config.MaxAllowedSize)
		}
	}

	results, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(url.Host),
		Key:    aws.String(url.Path),
	})
	if err != nil {
		return nil, err
	}

	defer results.Body.Close()

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, results.Body); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func init() {
	RegisterSource(ImageSourceTypeS3, NewS3ImageSource)
}
