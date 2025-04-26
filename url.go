package main

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)

func generatePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s3Client)
	r, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Key:    &key,
		Bucket: &bucket,
	}, s3.WithPresignExpires(expireTime))
	if err != nil {
		return "", err
	}

	return r.URL, nil

}

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
	if video.VideoURL == nil {
		return video, nil
	}
	splittedURL := strings.Split(*video.VideoURL, ",")
	bucket := splittedURL[0]
	key := splittedURL[1]
	presignedURL, err := generatePresignedURL(cfg.s3Client, bucket, key, time.Minute)
	if err != nil {
		return database.Video{}, err
	}
	video.VideoURL = &presignedURL
	return video, nil
}
