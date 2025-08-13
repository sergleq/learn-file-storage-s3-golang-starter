package main

import (
	"context"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
	"github.com/google/uuid"
)

// VideoUploadRequest represents the video upload request
type VideoUploadRequest struct {
	VideoID   uuid.UUID
	UserID    uuid.UUID
	File      io.ReadCloser
	Header    *multipart.FileHeader
	MediaType string
}

// VideoUploadResponse represents the video upload response
type VideoUploadResponse struct {
	VideoID  uuid.UUID `json:"video_id"`
	VideoURL string    `json:"video_url"`
	Message  string    `json:"message"`
}

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	// Step 1: Setup request limits
	r.Body = http.MaxBytesReader(w, r.Body, MaxVideoUploadSize)

	// Step 2: Parse and validate video ID
	videoID, err := cfg.parseAndValidateVideoID(r)
	if err != nil {
		respondWithError(w, StatusBadRequest, "Invalid video ID", err)
		return
	}

	// Step 3: Authenticate user
	userID, err := cfg.authenticateUser(r)
	if err != nil {
		respondWithError(w, StatusUnauthorized, "Authentication failed", err)
		return
	}

	// Step 4: Get and authorize video access
	video, err := cfg.getAndAuthorizeVideo(videoID, userID)
	if err != nil {
		respondWithError(w, StatusUnauthorized, "Not authorized to update this video", err)
		return
	}

	// Step 5: Parse and validate uploaded file
	uploadReq, err := cfg.parseAndValidateUploadedFile(r)
	if err != nil {
		respondWithError(w, StatusBadRequest, "Invalid uploaded file", err)
		return
	}
	defer uploadReq.File.Close()

	// Step 6: Process video file
	processedVideoPath, err := cfg.processVideoFile(uploadReq)
	if err != nil {
		respondWithError(w, StatusInternalServerError, "Failed to process video", err)
		return
	}

	// Step 7: Upload to S3
	s3Key, err := cfg.uploadVideoToS3(processedVideoPath, uploadReq.MediaType)
	if err != nil {
		respondWithError(w, StatusBadGateway, "Failed to upload to S3", err)
		return
	}

	// Step 8: Update database
	err = cfg.updateVideoInDatabase(video, s3Key)
	if err != nil {
		respondWithError(w, StatusInternalServerError, "Failed to update database", err)
		return
	}

	// Step 9: Return success response
	response := VideoUploadResponse{
		VideoID:  videoID,
		VideoURL: fmt.Sprintf("https://%s/%s", cfg.s3CfDistribution, s3Key),
		Message:  "Video uploaded successfully",
	}
	respondWithJSON(w, StatusOK, response)
}

// parseAndValidateVideoID extracts and validates the video ID from the request
func (cfg *apiConfig) parseAndValidateVideoID(r *http.Request) (uuid.UUID, error) {
	videoIDString := r.PathValue("videoID")
	if videoIDString == "" {
		return uuid.Nil, NewValidationError("videoID", "video ID is required")
	}

	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		return uuid.Nil, NewValidationError("videoID", "invalid video ID format")
	}

	return videoID, nil
}

// authenticateUser authenticates the user from the request
func (cfg *apiConfig) authenticateUser(r *http.Request) (uuid.UUID, error) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		return uuid.Nil, NewAuthorizationError("couldn't find JWT token")
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		return uuid.Nil, NewAuthorizationError("invalid JWT token")
	}

	return userID, nil
}

// getAndAuthorizeVideo retrieves the video and checks user authorization
func (cfg *apiConfig) getAndAuthorizeVideo(videoID, userID uuid.UUID) (*database.Video, error) {
	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		return nil, NewFileProcessingError("database", "couldn't find video")
	}

	if video.UserID != userID {
		return nil, NewAuthorizationError("not authorized to update this video")
	}

	return &video, nil
}

// parseAndValidateUploadedFile parses and validates the uploaded video file
func (cfg *apiConfig) parseAndValidateUploadedFile(r *http.Request) (*VideoUploadRequest, error) {
	file, header, err := r.FormFile("video")
	if err != nil {
		return nil, NewValidationError("video", "unable to parse form file")
	}

	mediaType, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if err != nil {
		return nil, NewValidationError("video", "invalid content type")
	}

	if mediaType != VideoMP4Type {
		return nil, NewValidationError("video", "only MP4 videos are supported")
	}

	return &VideoUploadRequest{
		File:      file,
		Header:    header,
		MediaType: mediaType,
	}, nil
}

// processVideoFile processes the uploaded video file
func (cfg *apiConfig) processVideoFile(req *VideoUploadRequest) (string, error) {
	// Create temporary file
	tmpFile, err := os.CreateTemp("", "tubely-upload.mp4")
	if err != nil {
		return "", NewFileProcessingError("temp_file", "couldn't create temp file")
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Copy uploaded file to temp file
	if _, err := io.Copy(tmpFile, req.File); err != nil {
		return "", NewFileProcessingError("copy", "couldn't copy to temp file")
	}

	// Reset file pointer to beginning
	if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
		return "", NewFileProcessingError("seek", "couldn't seek temp file")
	}

	// Process video for fast start streaming
	fastStartFilePath, err := processVideoForFastStart(tmpFile.Name())
	if err != nil {
		return "", NewFileProcessingError("fast_start", "couldn't create fast start file")
	}

	return fastStartFilePath, nil
}

// uploadVideoToS3 uploads the processed video to S3
func (cfg *apiConfig) uploadVideoToS3(processedVideoPath, mediaType string) (string, error) {
	ctx := context.Background()
	// Open processed video file
	fastStartFile, err := os.Open(processedVideoPath)
	if err != nil {
		return "", NewFileProcessingError("open", "couldn't open fast start file")
	}
	defer os.Remove(processedVideoPath)
	defer fastStartFile.Close()

	// Generate S3 key
	prefix, err := getVideoAspectRatio(processedVideoPath)
	if err != nil {
		return "", NewFileProcessingError("aspect_ratio", "couldn't determine video aspect ratio")
	}

	s3Key := fmt.Sprintf("%s/%s%s", prefix, getRandomAssetsName(32), MP4Extension)

	// Upload to S3
	_, err = cfg.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &cfg.s3Bucket,
		Key:         &s3Key,
		Body:        fastStartFile,
		ContentType: &mediaType,
	})
	if err != nil {
		return "", NewS3Error("upload", "failed to upload video to S3")
	}

	return s3Key, nil
}

// updateVideoInDatabase updates the video record in the database
func (cfg *apiConfig) updateVideoInDatabase(video *database.Video, s3Key string) error {
	videoURL := fmt.Sprintf("https://%s/%s", cfg.s3CfDistribution, s3Key)
	video.VideoURL = &videoURL

	return cfg.db.UpdateVideo(*video)
}
