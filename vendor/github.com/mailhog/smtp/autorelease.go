package smtp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mailhog/MailHog-Server/config"
)

const mailHogURL = "http://192.168.254.104:8025"

// ReleaseConfig is an alias to preserve go package API
type ReleaseConfig config.OutgoingSMTP

// AutoRelease automatically release the message based on the smtp settings
func AutoRelease(mailID string) {
	log.Println("auto sending email", mailID)

	resp, err := ReleaseEmail(mailID)
	if err != nil {
		log.Println("releasing email error ", err, resp)
	}
}

func ReleaseEmail(mailID string) (*http.Response, error) {
	apiEndpoint := fmt.Sprintf("/api/v1/messages/%s/auto-release", mailID)

	req, err := NewRequest("POST", apiEndpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := Do(req, nil)
	if err != nil {
		return resp, err
	}

	defer resp.Body.Close()

	return resp, nil
}

func GetMail(mailID string) (map[string]interface{}, error) {
	apiEndpoint := fmt.Sprintf("/api/v1/messages/%s", mailID)
	req, err := NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		return nil, err
	}
	data := make(map[string]interface{})
	_, err = Do(req, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func Do(req *http.Request, v interface{}) (*http.Response, error) {
	client := &http.Client{Timeout: 20 * time.Second}
	httpResp, err := client.Do(req)

	if err != nil {
		fmt.Println("client do error")
		return nil, err
	}

	if v != nil {
		// Open a NewDecoder and defer closing the reader only if there is a provided interface to decode to
		defer httpResp.Body.Close()
		err = json.NewDecoder(httpResp.Body).Decode(v)
		// err = ffjson.NewDecoder().DecodeReader(httpResp.Body, v)
	}

	return httpResp, nil
}

// NewRequest creates an API request.
// A relative URL can be provided in urlStr, in which case it is resolved relative to the baseURL of the Client.
// If specified, the value pointed to by body is JSON encoded and included as the request body.
func NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	// Relative URLs should be specified without a preceding slash since baseURL will have the trailing slash
	rel.Path = strings.TrimLeft(rel.Path, "/")

	u, err := url.Parse(mailHogURL)
	if err != nil {
		log.Fatal(err)
	}

	u = u.ResolveReference(rel)

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err = json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}
