package filedata

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	fileData "github.com/ente-io/museum/ente/filedata"
	"github.com/ente-io/stacktrace"
	log "github.com/sirupsen/logrus"
)

func (c *Controller) downloadObject(ctx context.Context, objectKey string, dc string) (fileData.S3FileMetadata, error) {
	var obj fileData.S3FileMetadata
	buff := &aws.WriteAtBuffer{}
	bucket := c.S3Config.GetBucket(dc)
	downloader := c.downloadManagerCache[dc]
	_, err := downloader.DownloadWithContext(ctx, buff, &s3.GetObjectInput{
		Bucket: bucket,
		Key:    &objectKey,
	})
	if err != nil {
		return obj, err
	}
	err = json.Unmarshal(buff.Bytes(), &obj)
	if err != nil {
		return obj, stacktrace.Propagate(err, "unmarshal failed")
	}
	return obj, nil
}

// uploadObject uploads the embedding object to the object store and returns the object size
func (c *Controller) uploadObject(obj fileData.S3FileMetadata, objectKey string, dc string) (int64, error) {
	embeddingObj, _ := json.Marshal(obj)
	s3Client := c.S3Config.GetS3Client(dc)
	s3Bucket := c.S3Config.GetBucket(dc)
	uploader := s3manager.NewUploaderWithClient(&s3Client)
	up := s3manager.UploadInput{
		Bucket: s3Bucket,
		Key:    &objectKey,
		Body:   bytes.NewReader(embeddingObj),
	}
	result, err := uploader.Upload(&up)
	if err != nil {
		log.Error(err)
		return -1, stacktrace.Propagate(err, "")
	}
	log.Infof("Uploaded to bucket %s", result.Location)
	return int64(len(embeddingObj)), nil
}
