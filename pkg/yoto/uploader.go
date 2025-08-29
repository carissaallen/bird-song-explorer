package yoto

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type AudioUploader struct {
	client      *Client
	maxAttempts int
}

type UploadURLResponse struct {
	Upload struct {
		UploadURL string `json:"uploadUrl"`
		UploadID  string `json:"uploadId"`
	} `json:"upload"`
}

type TranscodeResponse struct {
	Transcode struct {
		TranscodedSha256 string `json:"transcodedSha256"`
		TranscodedInfo   struct {
			Duration interface{} `json:"duration"` // Can be int or string
			FileSize interface{} `json:"fileSize"` // Can be int64 or string
			Channels interface{} `json:"channels"` // Can be int or string
			Format   string      `json:"format"`
			Metadata struct {
				Title string `json:"title"`
			} `json:"metadata"`
		} `json:"transcodedInfo"`
	} `json:"transcode"`
}

// Helper methods to get values as proper types
func (tr *TranscodeResponse) GetDuration() int {
	if tr.Transcode.TranscodedInfo.Duration == nil {
		return 0
	}
	switch v := tr.Transcode.TranscodedInfo.Duration.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case string:
		if val, err := strconv.Atoi(v); err == nil {
			return val
		}
	}
	return 0
}

func (tr *TranscodeResponse) GetFileSize() int64 {
	if tr.Transcode.TranscodedInfo.FileSize == nil {
		return 0
	}
	switch v := tr.Transcode.TranscodedInfo.FileSize.(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	case string:
		if val, err := strconv.ParseInt(v, 10, 64); err == nil {
			return val
		}
	}
	return 0
}

func (tr *TranscodeResponse) GetChannels() string {
	if tr.Transcode.TranscodedInfo.Channels == nil {
		return "stereo" // Default to stereo
	}
	switch v := tr.Transcode.TranscodedInfo.Channels.(type) {
	case float64:
		if int(v) == 1 {
			return "mono"
		}
		return "stereo"
	case int:
		if v == 1 {
			return "mono"
		}
		return "stereo"
	case string:
		// It might already be "stereo" or "mono"
		if v == "stereo" || v == "mono" {
			return v
		}
		// Or it might be a number as string
		if val, err := strconv.Atoi(v); err == nil {
			if val == 1 {
				return "mono"
			}
		}
		return "stereo"
	}
	return "stereo" // Default to stereo
}

type PlaylistContent struct {
	Title    string   `json:"title"`
	Content  Content  `json:"content"`
	Metadata Metadata `json:"metadata"`
}

type Content struct {
	Chapters []Chapter `json:"chapters"`
}

type Chapter struct {
	Key          string          `json:"key"`
	Title        string          `json:"title"`
	OverlayLabel string          `json:"overlayLabel"`
	Tracks       []PlaylistTrack `json:"tracks"`
	Display      Display         `json:"display"`
}

type PlaylistTrack struct {
	Key          string  `json:"key"`
	Title        string  `json:"title"`
	TrackURL     string  `json:"trackUrl"`
	Duration     int     `json:"duration"`
	FileSize     int64   `json:"fileSize"`
	Channels     string  `json:"channels"` // "stereo" or "mono"
	Format       string  `json:"format"`
	Type         string  `json:"type"`
	OverlayLabel string  `json:"overlayLabel"`
	Display      Display `json:"display"`
}

type Display struct {
	Icon16x16 string `json:"icon16x16"`
}

type Metadata struct {
	Media MediaInfo `json:"media"`
}

type MediaInfo struct {
	Duration         int     `json:"duration"`
	FileSize         int64   `json:"fileSize"`
	ReadableFileSize float64 `json:"readableFileSize"`
}

func NewAudioUploader(client *Client) *AudioUploader {
	return &AudioUploader{
		client:      client,
		maxAttempts: 30,
	}
}

// UploadAudioFile uploads a local audio file to Yoto
func (au *AudioUploader) UploadAudioFile(filePath string) (string, error) {
	if err := au.client.ensureAuthenticated(); err != nil {
		return "", fmt.Errorf("authentication failed: %w", err)
	}

	// Step 1: Get upload URL
	uploadURL, uploadID, err := au.getUploadURL()
	if err != nil {
		return "", fmt.Errorf("failed to get upload URL: %w", err)
	}

	// Step 2: Upload the file
	if err := au.uploadFile(uploadURL, filePath); err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Step 3: Wait for transcoding
	transcodedSha, err := au.waitForTranscoding(uploadID)
	if err != nil {
		return "", fmt.Errorf("transcoding failed: %w", err)
	}

	return transcodedSha, nil
}

// UploadAudioFromURL downloads and uploads audio from a URL
func (au *AudioUploader) UploadAudioFromURL(audioURL string, title string) (string, *TranscodeResponse, error) {
	// Download the audio file
	resp, err := http.Get(audioURL)
	if err != nil {
		return "", nil, fmt.Errorf("failed to download audio: %w", err)
	}
	defer resp.Body.Close()

	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read audio data: %w", err)
	}

	return au.UploadAudioData(audioData, title)
}

// UploadAudioData uploads raw audio data to Yoto
func (au *AudioUploader) UploadAudioData(audioData []byte, title string) (string, *TranscodeResponse, error) {
	uploadURL, uploadID, err := au.getUploadURL()
	if err != nil {
		return "", nil, fmt.Errorf("failed to get upload URL: %w", err)
	}

	// Upload the audio data
	req, err := http.NewRequest("PUT", uploadURL, bytes.NewReader(audioData))
	if err != nil {
		return "", nil, err
	}
	req.Header.Set("Content-Type", "audio/mpeg")

	uploadResp, err := au.client.httpClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("upload failed: %w", err)
	}
	defer uploadResp.Body.Close()

	if uploadResp.StatusCode != http.StatusOK && uploadResp.StatusCode != http.StatusCreated {
		return "", nil, fmt.Errorf("upload failed with status: %d", uploadResp.StatusCode)
	}

	// Wait for transcoding
	transcodeInfo, err := au.waitForTranscodingWithInfo(uploadID)
	if err != nil {
		return "", nil, fmt.Errorf("transcoding failed: %w", err)
	}

	return transcodeInfo.Transcode.TranscodedSha256, transcodeInfo, nil
}

func (au *AudioUploader) getUploadURL() (string, string, error) {
	url := fmt.Sprintf("%s/media/transcode/audio/uploadUrl", au.client.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", err
	}

	req.Header.Set("Authorization", "Bearer "+au.client.accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := au.client.httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("failed to get upload URL: %d - %s", resp.StatusCode, string(body))
	}

	var uploadResp UploadURLResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return "", "", err
	}

	return uploadResp.Upload.UploadURL, uploadResp.Upload.UploadID, nil
}

func (au *AudioUploader) uploadFile(uploadURL, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", uploadURL, file)
	if err != nil {
		return err
	}

	req.ContentLength = fileInfo.Size()
	req.Header.Set("Content-Type", "audio/mpeg")
	req.Header.Set("Content-Disposition", filepath.Base(filePath))

	resp, err := au.client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

func (au *AudioUploader) waitForTranscoding(uploadID string) (string, error) {
	for attempts := 0; attempts < au.maxAttempts; attempts++ {
		url := fmt.Sprintf("%s/media/upload/%s/transcoded?loudnorm=false", au.client.baseURL, uploadID)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return "", err
		}

		req.Header.Set("Authorization", "Bearer "+au.client.accessToken)
		req.Header.Set("Accept", "application/json")

		resp, err := au.client.httpClient.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var transcodeResp TranscodeResponse
			if err := json.NewDecoder(resp.Body).Decode(&transcodeResp); err != nil {
				return "", err
			}

			if transcodeResp.Transcode.TranscodedSha256 != "" {
				return transcodeResp.Transcode.TranscodedSha256, nil
			}
		}

		time.Sleep(500 * time.Millisecond)
	}

	return "", fmt.Errorf("transcoding timed out after %d attempts", au.maxAttempts)
}

func (au *AudioUploader) waitForTranscodingWithInfo(uploadID string) (*TranscodeResponse, error) {
	for attempts := 0; attempts < au.maxAttempts; attempts++ {
		url := fmt.Sprintf("%s/media/upload/%s/transcoded?loudnorm=false", au.client.baseURL, uploadID)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", "Bearer "+au.client.accessToken)
		req.Header.Set("Accept", "application/json")

		resp, err := au.client.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var transcodeResp TranscodeResponse
			if err := json.NewDecoder(resp.Body).Decode(&transcodeResp); err != nil {
				return nil, err
			}

			if transcodeResp.Transcode.TranscodedSha256 != "" {
				return &transcodeResp, nil
			}
		}

		time.Sleep(500 * time.Millisecond)
	}

	return nil, fmt.Errorf("transcoding timed out after %d attempts", au.maxAttempts)
}
