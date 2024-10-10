package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	BaseURL = "http://localhost:"
)

type Client struct {
	handle   *http.Client
	port     int
	apiToken string
}

// TODO define the structs
type Note struct{}

func New() (*Client, error) {
	newClient := Client{
		handle:   &http.Client{},
		port:     0,
		apiToken: "",
	}

	// TODO: find programatically the port if default is not valid
	newClient.port = 41184

	// read auth token from file and ignore new line
	auth_token, err := os.ReadFile("./.auth-token")
	if err != nil {
		fmt.Println(err)
	}

	newClient.apiToken = strings.TrimSpace(string(auth_token))

	return &newClient, nil
}

func (c *Client) GetNote(id string, fields string) (Note, error) {
	var note Note

	url := fmt.Sprintf("%s%d/notes/%s", BaseURL, c.port, id)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Print(err)
	}

	q := req.URL.Query()
	q.Add("fields", fields)
	q.Add("token", c.apiToken)
	req.URL.RawQuery = q.Encode()

	resp, err := c.handle.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	// TODO: remove this and only return the note
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(data))

	return note, err
}

func main() {
	newClient, _ := New()
	newClient.GetNote("<id>", "id,title,body")
}
