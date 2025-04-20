package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
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

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here

	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	// "thumbnail" should match the HTML form input name
	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Can't get extension from Content-Type header", err)
		return
	}

	if mediaType != "image/jpeg" && mediaType != "image/png" {
		respondWithError(w, http.StatusInternalServerError, "Forbidden upload type (only jpeg/png allowed)", nil)
		return
	}
	extensions, err := mime.ExtensionsByType(mediaType)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Can't get extension from Content-Type header", err)
		return
	}
	ext := extensions[0]

	videoMetaInfo, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Can't get video meta info from db", err)
		return
	}

	if videoMetaInfo.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "You don't have access to update this video", nil)
		return
	}

	randomBytes := make([]byte, 32)
	rand.Read(randomBytes)
	newThumbName := make([]byte, base64.RawURLEncoding.EncodedLen(len(randomBytes)))
	base64.RawURLEncoding.Encode(newThumbName, randomBytes)
	path := filepath.Join(cfg.assetsRoot, fmt.Sprint(string(newThumbName), ext))
	thumbnailFile, err := os.Create(path)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Can't create a file", err)
	}

	io.Copy(thumbnailFile, file)

	dataURL := fmt.Sprintf("http://localhost:%v/assets/%v%v", cfg.port, string(newThumbName), ext)

	videoMetaInfo.UpdatedAt = time.Now()
	videoMetaInfo.ThumbnailURL = &dataURL

	err = cfg.db.UpdateVideo(videoMetaInfo)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Can't update video info in db", err)
		return
	}

	respondWithJSON(w, http.StatusOK, videoMetaInfo)
}
