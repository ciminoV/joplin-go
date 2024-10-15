package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	BaseURL    = "http://localhost"
	MinPortNum = 41184
	MaxPortNum = 41194
)

/** Properties of a client */
type Client struct {
	handle   *http.Client
	port     int
	apiToken string
}

/** Properties of a note */
type Note struct {
	ID                   string  `json:"id"`
	ParentID             string  `json:"parent_id"`
	Title                string  `json:"title"`
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
	Type                 int     `json:"type_,omitempty"`
}

/** Multiple results are paginated with the following structure */
type notesResult struct {
	Items   []Note `json:"items"`
	HasMore bool   `json:"has_more"`
}

/** Create a new client. Find joplin port and retrieve the auth token */
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

	authTokenFile, err := os.ReadFile("./.auth-token")
	if err != nil {
		// TODO: get token programmaticaly if the file doesn't already exist
		retErr = err
		return nil, retErr
	}

	// ignore new line character
	newClient.apiToken = strings.TrimSpace(string(authTokenFile))

	return &newClient, nil
}

/** Retrieve a single note given an id and a string of fields */
func (c *Client) GetNote(id string, fields string) (Note, error) {
	var note Note

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s:%d/notes/%s", BaseURL, c.port, id), nil)
	if err != nil {
		log.Print(err)
	}

	q := req.URL.Query()
	q.Add("fields", fields)
	q.Add("token", c.apiToken)
	req.URL.RawQuery = q.Encode()

	// TODO: more verbose errors
	resp, err := c.handle.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal([]byte(data), &note)

	return note, err
}

/** Retrieve all the notes in a given order */
func (c *Client) GetAllNotes(fields string, order_by string, order_dir string) ([]Note, error) {
	var result notesResult
	var notes []Note

	page_num := 1

	for {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s:%d/notes/", BaseURL, c.port), nil)
		if err != nil {
			log.Print(err)
			return notes, err
		}

		q := req.URL.Query()
		if len(order_by) > 0 {
			q.Add("order_by", order_by)
		}
		if len(order_dir) > 0 {
			q.Add("order_dir", strings.ToUpper(order_dir))
		}
		q.Add("page", strconv.Itoa(page_num))
		q.Add("fields", fields)
		q.Add("token", c.apiToken)
		req.URL.RawQuery = q.Encode()

		// TODO: more verbose errors
		resp, err := c.handle.Do(req)
		if err != nil {
			log.Fatal(err)
			return notes, err
		}

		defer resp.Body.Close()

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
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

/** Create a new note with a given format (markdown or html) */
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
		log.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s:%d/notes/", BaseURL, c.port), bytes.NewReader(jsonData))
	if err != nil {
		log.Fatal(err)
	}

	q := req.URL.Query()
	q.Add("token", c.apiToken)
	req.URL.RawQuery = q.Encode()

	resp, err := c.handle.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	// Return the new note
	new_note, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal([]byte(new_note), &note)

	return note, err
}

func main() {
	newClient, newErr := New()
	if newErr != nil {
		fmt.Print("Error in creating new client: ", newErr)
	}
	// note1, _ := newClient.GetNote("<id>", "id,title,body,updated_time,is_conflict")
	// fmt.Printf("%d", note1.Body)
	notes, _ := newClient.GetAllNotes("id,title,body", "title", "asc")
	for _, el := range notes {
		fmt.Printf("%s\n", el.Title)
	}
	// newnote, _ := newClient.CreateNote("new note", "markdown", "Some note in **Markdown**")
	// fmt.Print(newnote.ID)
}
