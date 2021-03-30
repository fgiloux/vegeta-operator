package main

import (
	"context"
	"flag"
	"log"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	VContentType = "application/octet-stream"
	FileDir      = "/results/"
)

type Caller struct {
	endpoint        string
	bucketName      string
	region          string
	accessKeyID     string
	secretAccessKey string
	secure          bool
	minioClient     *minio.Client
}

var caller Caller

func init() {
	var host, port string
	// Read parameters
	if host = strings.TrimSpace(os.Getenv("BUCKET_HOST")); host == "" {
		log.Fatalln("S3 endpoint needs to be specified. Please set the BUCKET_HOST environment variable")
	}
	if port = strings.TrimSpace(os.Getenv("BUCKET_PORT")); port == "" {
		port = "443"
	}
	caller.endpoint = host + ":" + port
	if caller.bucketName = strings.TrimSpace(os.Getenv("BUCKET_NAME")); caller.bucketName == "" {
		log.Fatalln("S3 bucket name needs to be specified. Please set the BUCKET_NAME environment variable")
	}
	caller.region = strings.TrimSpace(os.Getenv("BUCKET_REGION"))
	caller.accessKeyID = strings.TrimSpace(os.Getenv("AWS_ACCESS_KEY_ID"))
	caller.secretAccessKey = strings.TrimSpace(os.Getenv("AWS_SECRET_ACCESS_KEY"))
	if secString := strings.TrimSpace(os.Getenv("BUCKET_SSL")); secString == "" {
		caller.secure = false
	} else {
		if sec, err := strconv.ParseBool(secString); err != nil {
			log.Fatalln("S3_SECURE environment variable needs to be a boolean: true/false")
		} else {
			caller.secure = sec
		}
	}

	// Initialize minio client object.
	var err error
	if caller.minioClient, err = minio.New(caller.endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(caller.accessKeyID, caller.secretAccessKey, ""),
		Secure: caller.secure,
	}); err != nil {
		log.Fatalln("Unable to initialize the S3 client", err)
	}
}

func (c Caller) Upload(ctx context.Context, filePath string) {
	if c.region == "" {
		log.Printf("Region not specified")
		exists, err := c.minioClient.BucketExists(ctx, c.bucketName)
		if err != nil {
			log.Fatalln(err)
		} else if exists {
			log.Printf("The bucket already exists %s. Processing further...\n", c.bucketName)
		} else {
			log.Fatalln("The bucket does not exist. Please either create it upfront or specify the environment variable S3_REGION for its creation")
		}
	} else {
		err := c.minioClient.MakeBucket(ctx, c.bucketName, minio.MakeBucketOptions{Region: c.region})
		if err != nil {
			// Check to see if the bucket already exists
			exists, errBucketExists := c.minioClient.BucketExists(ctx, c.bucketName)
			if errBucketExists == nil && exists {
				log.Printf("The bucket already exists %s. Processing further...\n", c.bucketName)
			} else {
				log.Fatalln("Unable to create bucket", err)
			}
		} else {
			log.Printf("Bucket successfully created %s\n", c.bucketName)
		}
	}
	_, fileName := path.Split(filePath)

	// Upload the file with FPutObject
	n, err := c.minioClient.FPutObject(ctx, c.bucketName, fileName, filePath, minio.PutObjectOptions{ContentType: VContentType})
	if err != nil {
		log.Fatalln("Unable to upload file", err)
	}

	log.Printf("Successfully uploaded %s of size %d\n", fileName, n)
}

func (c Caller) Download(ctx context.Context, objectPrefix string) {
	for object := range c.minioClient.ListObjects(ctx, c.bucketName, minio.ListObjectsOptions{Prefix: objectPrefix}) {
		if object.Err != nil {
			log.Fatalln("Unable to list objects", object.Err)
		}
		if err := c.minioClient.FGetObject(ctx, caller.bucketName, object.Key, FileDir+object.Key, minio.GetObjectOptions{}); err == nil {
			log.Printf("Downloaded: %s\n", object.Key)
		} else {
			log.Fatalln("Unable to download files", err)
		}
	}
}

func main() {
	ctx := context.Background()

	var cmd string
	flag.StringVar(&cmd, "command", "upload", "The action to be performed: either upload or download. It defaults to upload.")
	flag.Parse()

	switch cmd {
	case "download":
		// Only required for download
		if objectPrefix := strings.TrimSpace(os.Getenv("S3_OBJECT_PREFIX")); objectPrefix == "" {
			log.Fatalln("A prefix for the files to download needs to be specified. The environment variable S3_OBJECT_PREFIX has no value set")
		} else {
			caller.Download(ctx, objectPrefix)
		}
	case "upload":
		// Only required for upload
		if filePath := strings.TrimSpace(os.ExpandEnv(os.Getenv("S3_UPLOAD_FILE"))); filePath == "" {
			log.Fatalln("File to upload needs to be specified. The environment variable S3_UPLOAD_FILE has no value set")
		} else {
			caller.Upload(ctx, filePath)
		}
	default:
		log.Printf("The command %s provided is not valid. Accepted values are download or upload. If the command is not provided it will default to upload.\n", cmd)
	}
}
