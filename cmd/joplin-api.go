package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	BaseURL            = "http://localhost"
	AuthTokenPath      = "/.joplin-auth-token"
	MinPortNum         = 41184
	MaxPortNum         = 41194
	RetriesGetApiToken = 20
)

/** Properties of a client. */
type Client struct {
	handle   *http.Client
	port     int
	apiToken string
}

/** Properties of a note. */
type Note struct {
	ID                   string  `json:"id,omitempty"`
	ParentID             string  `json:"parent_id,omitempty"`
	Title                string  `json:"title,omitempty"`
	Body                 string  `json:"body,omitempty"`
	CreatedTime          int     `json:"created_time,omitempty"`
	UpdatedTime          int     `json:"updated_time,omitempty"`
	IsConflict           int     `json:"is_conflict,omitempty"`
	Latitude             float64 `json:"latitude,omitempty"`
	Longitude            float64 `json:"longitude,omitempty"`
	Altitude             float64 `json:"altitude,omitempty"`
	Author               string  `json:"author,omitempty"`
	SourceURL            string  `json:"source_url,omitempty"`
	IsTodo               int     `json:"is_todo,omitempty"`
	TodoDue              int     `json:"todo_due,omitempty"`
	TodoCompleted        int     `json:"todo_completed,omitempty"`
	Source               string  `json:"source,omitempty"`
	SourceApplication    string  `json:"source_application,omitempty"`
	ApplicationData      string  `json:"application_data,omitempty"`
	Order                float64 `json:"order,omitempty"`
	UserCreatedTime      int     `json:"user_created_time,omitempty"`
	UserUpdatedTime      int     `json:"user_updated_time,omitempty"`
	EncryptionCipherText string  `json:"encryption_cipher_text,omitempty"`
	EncryptionApplied    int     `json:"encryption_applied,omitempty"`
	MarkupLanguage       int     `json:"markup_language,omitempty"`
	IsShared             int     `json:"is_shared,omitempty"`
	ShareID              string  `json:"share_id,omitempty"`
	ConflictOriginalID   string  `json:"conflict_original_id,omitempty"`
	MasterKeyID          string  `json:"master_key_id,omitempty"`
	BodyHTML             string  `json:"body_html,omitempty"`
	BaseURL              string  `json:"base_url,omitempty"`
	ImageDataURL         string  `json:"image_data_url,omitempty"`
	CropRect             string  `json:"crop_rect,omitempty"`
}

/** Multiple results are paginated with the following structure. */
type notesResult struct {
	Items   []Note `json:"items"`
	HasMore bool   `json:"has_more"`
}

/** Create a new client. Find joplin port and retrieve the auth token. */
func New() (*Client, error) {
	var retErr error

	portFound := false

	newClient := Client{
		handle:   &http.Client{},
		port:     0,
		apiToken: "",
	}

	// Find the port on which the service is running
	for i := MinPortNum; i <= MaxPortNum; i++ {
		resp, err := http.Get(fmt.Sprintf("%s:%d/ping", BaseURL, i))
		if err != nil {
			retErr = err
			continue
		}

		defer resp.Body.Close()

		newClient.port = i

		portFound = true
		break
	}

	if !portFound {
		return nil, retErr
	}

	// Retrieve the authorisation token from file if any or request it programmatically
	homeDir, err := os.UserHomeDir()
	if err != nil {
		retErr = err
	}
	authDir := homeDir + AuthTokenPath

	if authTokenFile, err := os.ReadFile(authDir); err == nil {
		newClient.apiToken = strings.TrimSpace(string(authTokenFile))
	} else {
		var result struct {
			AuthToken string `json:"auth_token"`
			Status    string `json:"status"`
			ApiToken  string `json:"token,omitempty"`
		}

		// Get the auth token
		resp, err := http.Post(fmt.Sprintf("%s:%d/auth", BaseURL, newClient.port), "application/json", nil)
		if err != nil {
			retErr = err
			return nil, retErr
		}

		defer resp.Body.Close()

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			retErr = err
			return nil, retErr
		}

		json.Unmarshal([]byte(data), &result)

		// Wait for the user to accept
		retries := 0
		receivedApiToken := false

		for {
			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s:%d/auth/check", BaseURL, newClient.port), nil)
			if err != nil {
				retErr = err
				break
			}

			q := req.URL.Query()
			q.Add("auth_token", result.AuthToken)
			req.URL.RawQuery = q.Encode()

			resp, err := newClient.handle.Do(req)
			if err != nil {
				retErr = err
				break
			}

			defer resp.Body.Close()

			data, err := io.ReadAll(resp.Body)
			if err != nil {
				retErr = err
				break
			}

			json.Unmarshal([]byte(data), &result)

			if result.Status == "accepted" {
				receivedApiToken = true

				// Save the api token on a file
				newClient.apiToken = result.ApiToken
				if err := os.WriteFile(authDir, []byte(newClient.apiToken), 0666); err != nil {
					retErr = err
				}

				break
			} else if result.Status == "rejected" {
				err = errors.New("Api-token request rejected.")
				retErr = err

				break
			} else if result.Status == "waiting" {
				retries++

				if retries < RetriesGetApiToken {
					time.Sleep(time.Second)

					continue
				}

				retErr = fmt.Errorf("Api token could not get an answer from user.")

				break
			}
		}

		if !receivedApiToken {
			return nil, retErr
		}
	}

	return &newClient, nil
}

/** Retrieve a single note given an id and a string of fields. */
func (c *Client) GetNote(id string, fields string) (Note, error) {
	var note Note

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s:%d/notes/%s", BaseURL, c.port, id), nil)
	if err != nil {
		return note, err
	}

	q := req.URL.Query()
	q.Add("fields", fields)
	q.Add("token", c.apiToken)
	req.URL.RawQuery = q.Encode()

	resp, err := c.handle.Do(req)
	if err != nil {
		return note, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = fmt.Errorf("Error %d in retrieving note with ID %s", resp.StatusCode, id)
		return note, err
	}

	// Store the note
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return note, err
	}

	json.Unmarshal([]byte(data), &note)

	return note, err
}

/** Retrieve all the notes in a given order. */
func (c *Client) GetAllNotes(fields string, orderBy string, orderDir string) ([]Note, error) {
	var result notesResult
	var notes []Note

	page_num := 1

	for {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s:%d/notes/", BaseURL, c.port), nil)
		if err != nil {
			return notes, err
		}

		q := req.URL.Query()
		if len(orderBy) > 0 {
			q.Add("order_by", orderBy)
		}
		if len(orderDir) > 0 {
			q.Add("order_dir", strings.ToUpper(orderDir))
		}
		q.Add("page", strconv.Itoa(page_num))
		q.Add("fields", fields)
		q.Add("token", c.apiToken)
		req.URL.RawQuery = q.Encode()

		resp, err := c.handle.Do(req)
		if err != nil {
			return notes, err
		}

		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			err = fmt.Errorf("Error %d in retrieving the notes", resp.StatusCode)
			return notes, err
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return notes, err
		}

		json.Unmarshal([]byte(data), &result)

		// Save all the notes in the current page
		for _, note := range result.Items {
			notes = append(notes, note)
		}

		// Check if there are more notes
		if result.HasMore {
			page_num++
			continue
		}

		return notes, nil
	}
}

/** Create a new note with a given format (markdown or html). */
func (c *Client) CreateNote(title string, format string, body string) (Note, error) {
	var note Note
	var data map[string]string

	if format == "markdown" {
		data = map[string]string{
			"title": title,
			"body":  body,
		}
	} else if format == "html" {
		data = map[string]string{
			"title":     title,
			"body_html": body,
		}
	} else {
		return note, fmt.Errorf("Unknown note format.")
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return note, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s:%d/notes/", BaseURL, c.port), bytes.NewReader(jsonData))
	if err != nil {
		return note, err
	}

	q := req.URL.Query()
	q.Add("token", c.apiToken)
	req.URL.RawQuery = q.Encode()

	resp, err := c.handle.Do(req)
	if err != nil {
		return note, err
	}

	defer resp.Body.Close()

	// Get the new note
	new_note, err := io.ReadAll(resp.Body)
	if err != nil {
		return note, err
	}

	json.Unmarshal([]byte(new_note), &note)

	return note, nil
}

/** Update properties of an existing note with a given id. */
func (c *Client) UpdateNote(id string, properties []string) (Note, error) {
	var retNote Note

	// Transform the input vector of strings to a json string
	propertiesMap := make(map[string]string)
	for i := 0; i < len(properties); i += 2 {
		if i+1 < len(properties) {
			propertiesMap[properties[i]] = properties[i+1]
		}
	}
	propertiesJson, _ := json.Marshal(propertiesMap)

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s:%d/notes/%s", BaseURL, c.port, id), bytes.NewReader(propertiesJson))
	if err != nil {
		return retNote, err
	}

	q := req.URL.Query()
	q.Add("token", c.apiToken)
	req.URL.RawQuery = q.Encode()

	resp, err := c.handle.Do(req)
	if err != nil {
		return retNote, err
	}

	defer resp.Body.Close()

	// Return the updated note
	updatedNote, err := io.ReadAll(resp.Body)
	if err != nil {
		return retNote, err
	}

	json.Unmarshal([]byte(updatedNote), &retNote)

	return retNote, nil
}

/** Delete note with given ID if any. */
func (c *Client) DeleteNote(id string, permanent bool) (string, error) {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s:%d/notes/%s", BaseURL, c.port, id), nil)
	if err != nil {
		return "", err
	}

	q := req.URL.Query()
	if permanent {
		q.Add("permanent", "1")
	}
	q.Add("token", c.apiToken)
	req.URL.RawQuery = q.Encode()

	resp, err := c.handle.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	return id, nil
}
