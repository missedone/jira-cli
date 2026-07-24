package jira

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetIssueAttachment(t *testing.T) {
	tests := []struct {
		name    string
		version string
	}{
		{name: "v3", version: "3"},
		{name: "v2", version: "2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, fmt.Sprintf("/rest/api/%s/issue/TEST-1?fields=attachment", tt.version), r.URL.RequestURI())
				assert.Equal(t, "application/json", r.Header.Get("Accept"))
				_, _ = w.Write([]byte(`{"fields":{"attachment":[{"id":"10001","filename":"report.pdf"}]}}`))
			}))
			defer server.Close()

			client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))
			var (
				attachment *Attachment
				err        error
			)
			if tt.version == "2" {
				attachment, err = client.GetIssueAttachmentV2("TEST-1", "10001")
			} else {
				attachment, err = client.GetIssueAttachment("TEST-1", "10001")
			}

			require.NoError(t, err)
			assert.Equal(t, &Attachment{ID: "10001", Filename: "report.pdf"}, attachment)
		})
	}
}

func TestGetIssueAttachmentNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"fields":{"attachment":[]}}`))
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))
	attachment, err := client.GetIssueAttachment("TEST-1", "10001")

	assert.Nil(t, attachment)
	assert.EqualError(t, err, `jira: attachment "10001" not found on issue "TEST-1"`)
}

func TestDownloadAttachment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/3/attachment/content/10001", r.URL.Path)
		_, _ = w.Write([]byte("attachment contents"))
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))
	var dst bytes.Buffer

	require.NoError(t, client.DownloadAttachment("10001", &dst))
	assert.Equal(t, "attachment contents", dst.String())
}

func TestDownloadAttachmentUnexpectedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errorMessages":["Attachment not found"]}`))
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))
	err := client.DownloadAttachment("bad", &bytes.Buffer{})

	var unexpected *ErrUnexpectedResponse
	require.ErrorAs(t, err, &unexpected)
	assert.Equal(t, http.StatusNotFound, unexpected.StatusCode)
}
