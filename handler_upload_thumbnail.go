package main

import (
	"fmt"
	"io"
	"net/http"
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

	mediaType := header.Header.Get("Content-Type")
	imgData, err := io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Problem with loading image data using ReadAll", err)
	}
	videoMetaInfo, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Can't get video meta info from db", err)
	}

	if videoMetaInfo.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "You don't have access to update this video", nil)
		return
	}

	newThumbnail := thumbnail{
		data:      imgData,
		mediaType: mediaType,
	}

	videoThumbnails[videoID] = newThumbnail

	newThumbnailURL := fmt.Sprintf("http://localhost:%v/api/thumbnails/%v", cfg.port, videoID)

	videoMetaInfo.UpdatedAt = time.Now()
	videoMetaInfo.ThumbnailURL = &newThumbnailURL

	err = cfg.db.UpdateVideo(videoMetaInfo)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Can't update video info in db", err)
		return
	}

	respondWithJSON(w, http.StatusOK, videoMetaInfo)
}
