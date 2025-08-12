package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func getAssetPath(videoID string, mediaType string) string {
	ext := mediaTypeToExt(mediaType)
	return fmt.Sprintf("%s%s", videoID, ext)
}

func (cfg apiConfig) getAssetDiskPath(assetPath string) string {
	return filepath.Join(cfg.assetsRoot, assetPath)
}

func (cfg apiConfig) getAssetURL(assetPath string) string {
	return fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, assetPath)
}

func mediaTypeToExt(mediaType string) string {
	parts := strings.Split(mediaType, "/")
	if len(parts) != 2 {
		return ".bin"
	}
	return "." + parts[1]
}

func getRandomAssetsName(numByte int) string {
	bt := make([]byte, numByte)
	rand.Read(bt)
	return base64.RawURLEncoding.EncodeToString(bt)
}

func getVideoAspectRatio(filePath string) (string, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)

	var bufer bytes.Buffer
	cmd.Stdout = &bufer

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffprobe: %v", err)
	}

	type aspectRatio struct {
		Streams []struct {
			AR string `json:"display_aspect_ratio"`
		} `json:"streams"`
	}
	var aspRat aspectRatio

	if err := json.Unmarshal(bufer.Bytes(), &aspRat); err != nil {
		return "", fmt.Errorf("unmarshal ffprobe json: %w", err)
	}

	switch aspRat.Streams[0].AR {
	case "16:9":
		return "landscape", nil
	case "9:16":
		return "portrait", nil
	default:
		return "other", nil
	}
}

func processVideoForFastStart(filePath string) (string, error) {
	newPath := filePath + ".processing"

	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", newPath)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffprobe: %v", err)
	}

	return newPath, nil
}
