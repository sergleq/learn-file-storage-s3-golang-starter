package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)

func generatePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s3Client)

	presignedRequest, err := presignClient.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}, s3.WithPresignExpires(expireTime))

	if err != nil {
		return "", err
	}

	return presignedRequest.URL, nil
}

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
	if video.VideoURL == nil {
		return video, nil
	}

	// Split the video.VideoURL on comma to get bucket and key
	parts := strings.Split(*video.VideoURL, ",")
	if len(parts) != 2 {
		return video, fmt.Errorf("invalid video URL format: expected 'bucket,key', got '%s'", *video.VideoURL)
	}

	bucket := parts[0]
	key := parts[1]

	// Generate presigned URL with configured expiration
	presignedURL, err := generatePresignedURL(cfg.s3Client, bucket, key, PresignedURLExpiration)
	if err != nil {
		return video, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	// Update the video with the presigned URL
	video.VideoURL = &presignedURL

	return video, nil
}
