package utils

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"mime/multipart"
	"net/url"
	"sort"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryUploader struct {
	cld       *cloudinary.Cloudinary
	apiKey    string
	apiSecret string
	cloudName string
}

func NewCloudinaryUploader(cloudinaryURL string) (*CloudinaryUploader, error) {
	if cloudinaryURL == "" {
		return nil, errors.New("CLOUDINARY_URL is required")
	}

	cld, err := cloudinary.NewFromURL(cloudinaryURL)
	if err != nil {
		return nil, err
	}

	parsedURL, err := url.Parse(cloudinaryURL)
	if err != nil {
		return nil, errors.New("invalid CLOUDINARY_URL")
	}

	apiKey := parsedURL.User.Username()
	apiSecret, _ := parsedURL.User.Password()
	cloudName := parsedURL.Host

	return &CloudinaryUploader{
		cld:       cld,
		apiKey:    apiKey,
		apiSecret: apiSecret,
		cloudName: cloudName,
	}, nil
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

func (c *CloudinaryUploader) DeleteImage(ctx context.Context, publicID string) error {
	if publicID == "" {
		return errors.New("publicID is required")
	}

	params := uploader.DestroyParams{
		PublicID:     publicID,
		ResourceType: "image",
	}

	_, err := c.cld.Upload.Destroy(ctx, params)
	return err
}

func (c *CloudinaryUploader) GenerateUploadSignature(params map[string]interface{}) (string, error) {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var toSignStr string
	for i, k := range keys {
		val := fmt.Sprintf("%v", params[k])
		if i > 0 {
			toSignStr += "&"
		}
		toSignStr += fmt.Sprintf("%s=%s", k, val)
	}

	toSignStr += c.apiSecret

	hash := sha1.New()
	hash.Write([]byte(toSignStr))
	signature := hex.EncodeToString(hash.Sum(nil))

	return signature, nil
}

func (c *CloudinaryUploader) GetAPIKey() string {
	return c.apiKey
}

func (c *CloudinaryUploader) GetCloudName() string {
	return c.cloudName
}
