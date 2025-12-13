package utils

import (
	"context"
	"errors"
	"mime/multipart"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryUploader struct {
	cld *cloudinary.Cloudinary
}

func NewCloudinaryUploader() (*CloudinaryUploader, error) {
	cld, err := cloudinary.New()
	if err != nil {
		return nil, err
	}
	return &CloudinaryUploader{cld: cld}, nil
}

func (c *CloudinaryUploader) UploadImage(
	ctx context.Context,
	file multipart.File,
	folder string,
	format string, 
) (string, error) {

	if file == nil {
		return "", errors.New("invalid file")
	}

	params := uploader.UploadParams{
		Folder:       folder,
		ResourceType: "image",
	}

	// Force output format if provided
	if format != "" {
		params.Format = format
	}

	result, err := c.cld.Upload.Upload(ctx, file, params)
	if err != nil {
		return "", err
	}

	return result.SecureURL, nil
}
