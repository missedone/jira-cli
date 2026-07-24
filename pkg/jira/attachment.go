package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Attachment holds Jira attachment metadata.
type Attachment struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
}

// GetIssueAttachment fetches attachment metadata from an issue using the v3 API.
func (c *Client) GetIssueAttachment(issueKey, attachmentID string) (*Attachment, error) {
	return c.getIssueAttachment(issueKey, attachmentID, apiVersion3)
}

// GetIssueAttachmentV2 fetches attachment metadata from an issue using the v2 API.
func (c *Client) GetIssueAttachmentV2(issueKey, attachmentID string) (*Attachment, error) {
	return c.getIssueAttachment(issueKey, attachmentID, apiVersion2)
}

func (c *Client) getIssueAttachment(issueKey, attachmentID, ver string) (*Attachment, error) {
	path := fmt.Sprintf("/issue/%s?fields=attachment", url.PathEscape(issueKey))

	var (
		res *http.Response
		err error
	)
	if ver == apiVersion2 {
		res, err = c.GetV2(context.Background(), path, Header{"Accept": "application/json"})
	} else {
		res, err = c.Get(context.Background(), path, Header{"Accept": "application/json"})
	}
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, ErrEmptyResponse
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return nil, formatUnexpectedResponse(res)
	}

	var out struct {
		Fields struct {
			Attachments []*Attachment `json:"attachment"`
		} `json:"fields"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}

	for _, attachment := range out.Fields.Attachments {
		if attachment != nil && attachment.ID == attachmentID {
			return attachment, nil
		}
	}

	return nil, fmt.Errorf("jira: attachment %q not found on issue %q", attachmentID, issueKey)
}

// DownloadAttachment writes attachment content fetched from the v3 API to dst.
func (c *Client) DownloadAttachment(attachmentID string, dst io.Writer) error {
	return c.downloadAttachment(attachmentID, dst, apiVersion3)
}

// DownloadAttachmentV2 writes attachment content fetched from the v2 API to dst.
func (c *Client) DownloadAttachmentV2(attachmentID string, dst io.Writer) error {
	return c.downloadAttachment(attachmentID, dst, apiVersion2)
}

func (c *Client) downloadAttachment(attachmentID string, dst io.Writer, ver string) error {
	path := fmt.Sprintf("/attachment/content/%s", url.PathEscape(attachmentID))

	var (
		res *http.Response
		err error
	)
	if ver == apiVersion2 {
		res, err = c.GetV2(context.Background(), path, nil)
	} else {
		res, err = c.Get(context.Background(), path, nil)
	}
	if err != nil {
		return err
	}
	if res == nil {
		return ErrEmptyResponse
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return formatUnexpectedResponse(res)
	}

	_, err = io.Copy(dst, res.Body)
	return err
}
