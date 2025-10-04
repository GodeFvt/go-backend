package minio

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	helperModel "github.com/GodeFvt/go-backend/helper/models"
	"github.com/minio/minio-go/v7"
	minioLib "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	MINIO_DEFAULT_REGION = "ap-southeast-1"
)

type Client struct {
	MinioClient    *minioLib.Client
	MinioEndPoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioSSL       bool
	Region         string
}

func NewMinio(endpoint, access, secret string, ssl bool, region string) (*Client, error) {
	cli, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(access, secret, ""),
		Secure: ssl,
		Region: region,
	})
	if err != nil {
		return nil, err
	}
	return &Client{
		MinioClient:    cli,
		MinioEndPoint:  endpoint,
		MinioAccessKey: access,
		MinioSecretKey: secret,
		MinioSSL:       ssl,
		Region:         region,
	}, nil
}

func (c *Client) GetMinioURI() string {
	var minioEndURI string
	if c.MinioSSL {
		minioEndURI = "https://"
	} else {
		minioEndURI = "http://"
	}
	minioEndURI = minioEndURI + c.MinioEndPoint
	return minioEndURI
}

func CreateImageFile(img image.Image, format string, filename string, desc string) error {
	file, err := os.Create(desc)
	if err != nil {
		return err
	}
	switch format {
	case "jpg":
		if err := jpeg.Encode(file, img, nil); err != nil {
			return err
		}
		break
	case "jpeg":
		if err := jpeg.Encode(file, img, nil); err != nil {
			return err
		}
		break
	case "png":
		if err := png.Encode(file, img); err != nil {
			return err
		}
		break
	}
	defer file.Close()

	return nil
}

/*
This function will generate path filename for upload minio
format

	YYYYMMDD_id_[0-9]{10}_extension

	- YYYY ปีแบบคริสศักราช
	- MM เดือน
	- DD วันที่
	- id string
	- [0-9]{10} เลข generate 0-9 จำนวน 10 ตัว
	- extension

example

	20200413-9026362617-example.jpg
*/
func generateObjectName(foldername string, id string, extension string) string {
	var date = strings.ReplaceAll(helperModel.NewDateFromTime(time.Now()).String(), "-", "")
	var generateNumber = strconv.Itoa(rand.Intn(1000000000))
	extension = strings.TrimPrefix(extension, ".")

	foldername = func() string {
		lastString := []rune(foldername)
		if string(lastString) == "/" {
			return foldername
		}
		return foldername + "/"
	}()

	return fmt.Sprintf(foldername+"%s_%s_%s.%s", date, id, generateNumber, extension)
}

func GenerateObjectName(foldername string, id string, filename string) string {
	return generateObjectName(foldername, id, filename)
}

func (c *Client) GenerateObjectName(foldername string, id string, filename string) string {
	return generateObjectName(foldername, id, filename)
}

func (c *Client) GetClient() *minioLib.Client {
	return c.MinioClient
}

func (c *Client) GetEndPoint() string {
	return c.MinioEndPoint
}

func (c *Client) UploadMultipartFile(ctx context.Context, bucketName string, objectName string, file *multipart.FileHeader) (err error) {
	contentType := file.Header.Get("Content-Type")
	size := file.Size

	src, err := file.Open()
	if err != nil {
		return err
	}

	defer src.Close()

	if _, err = c.GetClient().PutObject(ctx, bucketName, objectName, src, size, minio.PutObjectOptions{ContentType: contentType}); err != nil {
		return err
	}
	return nil
}

func (c *Client) UploadFileWithReader(ctx context.Context, bucketName string, objectName string, reader io.Reader, size int64, contentType string, contentEncoding string) (err error) {
	if _, err = c.GetClient().PutObject(ctx, bucketName, objectName, reader, size, minio.PutObjectOptions{ContentType: contentType, ContentEncoding: contentEncoding}); err != nil {
		return err
	}
	return nil
}

func (c *Client) UploadFromFile(ctx context.Context, bucketName string, foldername string, pathFile string, filename string) error {
	objectName := foldername + "/" + filename

	src, err := os.Open(pathFile)
	if err != nil {
		return err
	}

	defer src.Close()

	if _, err := c.GetClient().FPutObject(ctx, bucketName, objectName, pathFile, minio.PutObjectOptions{}); err != nil {
		return err
	}
	return nil
}

func (c *Client) UploadFromFilePDF(ctx context.Context, bucketName string, foldername string, pathFile string, filename string) error {
	objectName := foldername + "/" + filename

	src, err := os.Open(pathFile)
	if err != nil {
		return err
	}

	defer src.Close()

	if _, err := c.GetClient().FPutObject(ctx, bucketName, objectName, pathFile, minio.PutObjectOptions{ContentType: "application/pdf", ContentEncoding: "UTF-8"}); err != nil {
		return err
	}
	return nil
}

func (c *Client) RemoveObject(ctx context.Context, bucketName string, objectName string) error {
	if err := c.GetClient().RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{}); err != nil {
		return err
	}
	return nil
}

func (c *Client) CreateBucket(ctx context.Context, bucketName string, region string) error {
	if region == "" {
		region = MINIO_DEFAULT_REGION
	}

	if err := c.GetClient().MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: region}); err != nil {
		return err
	}
	return nil
}

func (c *Client) ExistBucket(ctx context.Context, bucketName string) (bool, error) {
	exists, err := c.GetClient().BucketExists(ctx, bucketName)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (c *Client) SetBucketPublicPolicy(ctx context.Context, bucketName string) error {
	var buf bytes.Buffer
	bu, err := os.ReadFile("./policy/policy_public.json")
	if err != nil {
		return err
	}
	t, err := template.New("policy").Parse(string(bu))
	if err != nil {
		return err
	}

	if err := t.Execute(&buf, bucketName); err != nil {
		return err
	}

	policy := buf.String()

	if err := c.GetClient().SetBucketPolicy(ctx, bucketName, policy); err != nil {
		return err
	}
	log.Println("create bucket with policy success")
	fmt.Println(policy)
	return nil
}

func (c *Client) GetObjectnameFromURL(link string) (string, string) {
	var bucket string
	var pathImageRegex = regexp.MustCompile(`\/[a-z\-]+\/`)

	uri, err := url.Parse(link)
	if err != nil {
		return bucket, link
	}

	objectName := uri.Path
	if objectName != "" {
		if pathImageRegex.MatchString(objectName) && uri.Scheme != "" {
			loc := pathImageRegex.FindStringIndex(objectName)
			if len(loc) > 0 {
				bucket = strings.Trim(objectName[loc[0]:loc[1]], "/")
				objectName = objectName[loc[1]:]
			}
		}
	}
	if strings.Contains(objectName, "?") {
		spl := strings.Split(objectName, "?")
		objectName = spl[0]
	}
	if objectName == "" {
		return bucket, link
	}

	return bucket, strings.Trim(objectName, "/")
}
