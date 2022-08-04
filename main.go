package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	s3Client *minio.Client
)

func init() {
	// load env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	var err error

	s3Client, err = minio.New(os.Getenv("SPACES_ENDPOINT"), &minio.Options{
		Creds:  credentials.NewStaticV4(os.Getenv("SPACES_ACCESS_KEY_ID"), os.Getenv("SPACES_SECRET_ACCESS_KEY"), ""),
		Secure: true,
	})
	if err != nil {
		log.Fatalln(err)
	}

	// example uploading series of images
	startingIndex := 0
	endingIndex := 1

	for i := startingIndex; i <= endingIndex; i++ {
		filePath := fmt.Sprintf("./data/%d.png", i)
		err = upload(filePath, "images")
		if err != nil {
			log.Println("Error while uploading:", err)
		}
	}
}

func upload(path string, s3Path string) error {
	// open file
	object, err := os.Open(path)
	if err != nil {
		return err
	}
	defer object.Close()

	// get file info structure
	objectStat, err := object.Stat()
	if err != nil {
		return err
	}

	// get file content
	contentType, err := getFileContentType(object)
	if err != nil {
		return err
	}

	// object options
	opts := minio.PutObjectOptions{
		ContentType:  contentType,
		CacheControl: "max-age=31536000",
		UserMetadata: map[string]string{
			"x-amz-acl": "public-read", // NOTE: denotes public record
		},
	}

	// format s3 path
	if len(s3Path) > 0 {
		s3Path = fmt.Sprintf("%s/%s", s3Path, objectStat.Name())
	} else {
		s3Path = objectStat.Name()
	}

	n, err := s3Client.FPutObject(context.Background(), os.Getenv("SPACES_BUCKET_NAME"), s3Path, path, opts)
	if err != nil {
		return err
	}

	log.Println("Uploaded", s3Path, " of size: ", n, "Successfully.")
	return nil
}

func getFileContentType(ouput *os.File) (string, error) {
	// to sniff the content type only the first
	// 512 bytes are used.
	buf := make([]byte, 512)

	_, err := ouput.Read(buf)
	if err != nil {
		return "", err
	}

	// the function that actually does the trick
	contentType := http.DetectContentType(buf)

	return contentType, nil
}
