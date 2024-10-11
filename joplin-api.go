package main

import (
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
	BaseURL = "http://localhost"
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

/** All API calls that return multiple results will be paginated and will return the following structure */
type notesResult struct {
	Items   []Note `json:"items"`
	HasMore bool   `json:"has_more"`
}

/** Create a new client. Find joplin port and retrieve the auth token */
func New() (*Client, error) {
	newClient := Client{
		handle:   &http.Client{},
		port:     0,
		apiToken: "",
	}

	// TODO: find programatically the port if default is not valid
	newClient.port = 41184

	// TODO: get token programmaticaly if the file doesn't already exist
	auth_token, err := os.ReadFile("./.auth-token")
	if err != nil {
		fmt.Println(err)
	}

	// ignore new line character
	newClient.apiToken = strings.TrimSpace(string(auth_token))

	return &newClient, err
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

		// Save all the notes
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

func main() {
	newClient, _ := New()
	// note1, _ := newClient.GetNote("<id>", "id,title,body,updated_time,is_conflict")
	// fmt.Printf("%d", note1.Body)
	notes, _ := newClient.GetAllNotes("id,title,body", "title", "asc")
	for _, el := range notes {
		fmt.Printf("%s\n", el.Title)
	}
}
