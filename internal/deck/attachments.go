package deck

import (
	"context"
	"fmt"
)

func (c *Client) ListAttachments(ctx context.Context, boardID, stackID, cardID int64) ([]Attachment, error) {
	var attachments []Attachment
	err := c.doAppJSON(ctx, "GET", fmt.Sprintf("/cards/%d/attachments", cardID), nil, &attachments)
	return attachments, err
}

func (c *Client) UploadAttachment(ctx context.Context, boardID, stackID, cardID int64, filePath string) (Attachment, error) {
	var attachment Attachment
	err := c.doMultipart(ctx, "POST", c.appURL(fmt.Sprintf("/cards/%d/attachment", cardID)), map[string]string{"type": "file"}, "file", filePath, &attachment)
	return attachment, err
}

func (c *Client) DeleteAttachment(ctx context.Context, boardID, stackID, cardID, attachmentID int64) error {
	ref, err := c.attachmentRef(ctx, boardID, stackID, cardID, attachmentID)
	if err != nil {
		return err
	}
	return c.doAppJSON(ctx, "DELETE", fmt.Sprintf("/cards/%d/attachment/%s", cardID, ref), nil, nil)
}

func (c *Client) RestoreAttachment(ctx context.Context, boardID, stackID, cardID, attachmentID int64) (Attachment, error) {
	ref, err := c.attachmentRef(ctx, boardID, stackID, cardID, attachmentID)
	if err != nil {
		return Attachment{}, err
	}
	var attachment Attachment
	err = c.doAppJSON(ctx, "GET", fmt.Sprintf("/cards/%d/attachment/%s/restore", cardID, ref), nil, &attachment)
	return attachment, err
}

func (c *Client) DownloadAttachment(ctx context.Context, boardID, stackID, cardID, attachmentID int64, outPath string) error {
	ref, err := c.attachmentRef(ctx, boardID, stackID, cardID, attachmentID)
	if err != nil {
		return err
	}
	return c.DownloadAppFile(ctx, fmt.Sprintf("/cards/%d/attachment/%s", cardID, ref), outPath)
}

func (c *Client) attachmentRef(ctx context.Context, boardID, stackID, cardID, attachmentID int64) (string, error) {
	attachments, err := c.ListAttachments(ctx, boardID, stackID, cardID)
	if err != nil {
		return "", err
	}
	for _, attachment := range attachments {
		if attachment.ID == attachmentID {
			attachmentType := attachment.Type
			if attachmentType == "" {
				attachmentType = "deck_file"
			}
			return fmt.Sprintf("%s:%d", attachmentType, attachmentID), nil
		}
	}
	return "", fmt.Errorf("attachment %d not found", attachmentID)
}
