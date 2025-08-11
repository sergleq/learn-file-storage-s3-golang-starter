package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	maxMemory := 10 << 20
	if err := r.ParseMultipartForm(int64(maxMemory)); err != nil {
		respondWithError(w, http.StatusInternalServerError, "parsing is fail", err)
		return
	}

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")

	// byteImage, err := io.ReadAll(file)
	// if err != nil {
	// 	respondWithError(w, http.StatusInternalServerError, "Unable to read form file", err)
	// 	return
	// }

	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Видео недоступно", err)
		return
	}

	// thumb := thumbnail{
	// 	data:      byteImage,
	// 	mediaType: contentType,
	// }
	// videoThumbnails[videoID] = thumb
	// stringImage := base64.StdEncoding.EncodeToString(byteImage)

	ext, _ := mime.ExtensionsByType(contentType)
	if len(ext) == 0 {
		ext = []string{".bin"}
	}
	videoTypeStr := fmt.Sprintf("%s%s", videoIDString, ext[0])
	filePath := filepath.Join(cfg.assetsRoot, videoTypeStr)

	fileCreate, err := os.Create(filePath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "не удалось создать файл", err)
		return
	}
	if _, err := io.Copy(fileCreate, file); err != nil {
		respondWithError(w, http.StatusInternalServerError, "не удалось копировать файл", err)
		return
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	urlThumb := fmt.Sprintf("%s://%s/assets/%s", scheme, r.Host, videoTypeStr)
	// urlThumb := fmt.Sprintf("data:%s;base64,%s", contentType, stringImage)
	video.ThumbnailURL = &urlThumb

	if err := cfg.db.UpdateVideo(video); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to write database", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)
}
