package infrastructure

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/theboss/ajk-emoji/ajk-func/src/model"
)

// Storage represents storage
type Storage struct {
	bucketName string
	urlPrefix  string
	client     *s3.S3
}

// NewStorage returns instance storage
func NewStorage() *Storage {
	ep := os.Getenv("S3_ENDPOINT")
	config := aws.NewConfig().WithS3ForcePathStyle(true)
	if ep != "" {
		config = config.WithEndpoint(ep)
	}

	up := os.Getenv("S3_URL_PREFIX")
	bn := os.Getenv("S3_BUCKET_NAME")
	if up == "" {
		up = fmt.Sprintf(
			"https://s3-%s.amazonaws.com/%s",
			os.Getenv("AWS_DEFAULT_REGION"),
			bn,
		)
	}
	s := s3.New(session.New(), config)
	log.Printf("bucketName=%s urlPrefix=%s", bn, up)
	return &Storage{
		bucketName: bn,
		urlPrefix:  up,
		client:     s,
	}
}

// GetObjectURLPrefix returns urlPrefix
func (s *Storage) GetObjectURLPrefix() string {
	return s.urlPrefix
}

// Put puts object
func (s *Storage) Put(key string, b []byte) error {
	_, err := s.client.PutObject(&s3.PutObjectInput{
		Body:   bytes.NewReader(b),
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	return err
}

// PutImage puts image object
func (s *Storage) PutImage(img *model.Image) error {
	b, err := img.GetBytes()
	if err != nil {
		return err
	}
	return s.Put(img.GetFullName(), b)
}

// PutFile puts file object
func (s *Storage) PutFile(fpath, key string) error {
	f, err := os.Open(fpath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = s.client.PutObject(&s3.PutObjectInput{
		Body:   f,
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	return err
}

// FindByPrefix finds object by prefix key
func (s *Storage) FindByPrefix(prefix string) ([]string, error) {
	output, err := s.client.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(s.bucketName),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, err
	}
	var keys []string
	for _, item := range output.Contents {
		keys = append(keys, *item.Key)
	}
	return keys, nil
}

// Get gets object
func (s *Storage) Get(key string) (*model.StoreObject, error) {
	o, err := s.client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return &model.StoreObject{
		Body:          o.Body,
		ContentLength: *o.ContentLength,
	}, nil
}

// Head gets heads of object
func (s *Storage) Head(key string) (*s3.HeadObjectOutput, error) {
	o, err := s.client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return o, nil
}
