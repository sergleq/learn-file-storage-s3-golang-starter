package main

// HTTP status codes
const (
	StatusOK                  = 200
	StatusCreated             = 201
	StatusNoContent           = 204
	StatusBadRequest          = 400
	StatusUnauthorized        = 401
	StatusForbidden           = 403
	StatusNotFound            = 404
	StatusInternalServerError = 500
	StatusBadGateway          = 502
)

// File size limits
const (
	MaxVideoUploadSize = 1 << 30  // 1 GB
	MaxThumbnailSize   = 10 << 20 // 10 MB
)

// Time durations
const ()

// Media types
const (
	VideoMP4Type  = "video/mp4"
	ImageJPEGType = "image/jpeg"
	ImagePNGType  = "image/png"
)

// File extensions
const (
	MP4Extension  = ".mp4"
	JPEGExtension = ".jpg"
	PNGExtension  = ".png"
)

// S3 key prefixes
const (
	LandscapePrefix = "landscape"
	PortraitPrefix  = "portrait"
	OtherPrefix     = "other"
)

// Aspect ratios
const (
	AspectRatio16x9 = "16:9"
	AspectRatio9x16 = "9:16"
)
