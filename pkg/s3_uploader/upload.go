package s3_uploader

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"validator/v1/pkg/app_config"
)

type s3Uploader struct {
	appConfig  app_config.Config
	session    *session.Session
	service    *s3.S3
	group      *sync.WaitGroup
	isCorrect  bool
	baseFolder string
}

func NewS3Uploader(config app_config.Config) *s3Uploader {
	s3Uploader := s3Uploader{group: &sync.WaitGroup{}, appConfig: config}

	creds := credentials.NewStaticCredentialsFromCreds(credentials.Value{
		AccessKeyID:     config.S3Config.AccessKeyID,
		SecretAccessKey: config.S3Config.SecretAccessKey,
	})
	s3Uploader.session = session.Must(session.NewSession(
		&aws.Config{
			Credentials: creds,
			Endpoint:    &config.S3Config.EndpointUrl,
			Region:      &config.S3Config.Region,
			LogLevel:    nil,
		}))
	s3Uploader.service = s3.New(s3Uploader.session)
	return &s3Uploader
}

func (u *s3Uploader) generateS3FolderName(isCorrect bool) string {
	name := u.appConfig.CallbackTaskId + "/"
	if isCorrect {
		name += "__correct__/"
	}
	name += fmt.Sprintf(strings.Replace(u.appConfig.StudentConfig.StudentRepo, "/", "_", -1))
	return name
}

func (u *s3Uploader) UploadRepo(isCorrect bool) error {
	u.isCorrect = isCorrect
	u.baseFolder = u.generateS3FolderName(u.isCorrect)
	u.removeDirs()

	err := filepath.WalkDir(".", u.uploadDir)
	if err != nil {
		log.Printf("Can not upload repo to s3: %v", err)
		return err
	}
	return nil
}

func (u *s3Uploader) deleteDir(wg *sync.WaitGroup, path string) {
	defer func() {
		if wg != nil {
			wg.Done()
		}
	}()
	res, err := u.service.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(u.appConfig.S3Config.Bucket), Prefix: aws.String(path)})
	if err != nil {
		return
	}
	for _, object := range res.Contents {
		_, _ = u.service.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(u.appConfig.S3Config.Bucket),
			Key:    object.Key,
		})
	}
}

func (u *s3Uploader) removeDirs() {
	if u.isCorrect == true {
		wg := sync.WaitGroup{}
		wg.Add(2)
		go u.deleteDir(&wg, u.baseFolder)
		go u.deleteDir(&wg, u.generateS3FolderName(!u.isCorrect))
		wg.Wait()
	} else {
		u.deleteDir(nil, u.baseFolder)
	}
}

func (u *s3Uploader) uploadDir(path string, d os.DirEntry, err error) error {
	if d.IsDir() && needExcludeFolder(d.Name()) {
		return filepath.SkipDir
	}
	if d.IsDir() || d.Name() == "main" {
		return nil
	}
	err = u.uploadFile(path)
	if err != nil {
		return err
	}
	return nil
}

func (u *s3Uploader) uploadFile(originalPath string) error {
	f, err := os.Open(originalPath)
	defer func() { _ = f.Close() }()
	if err != nil {
		return err
	}
	reader := io.ReadSeeker(f)
	_, err = u.service.PutObject(&s3.PutObjectInput{
		ACL:    aws.String(s3.BucketCannedACLPublicRead),
		Body:   reader,
		Bucket: aws.String(u.appConfig.S3Config.Bucket),
		Key:    aws.String(filepath.Join(u.baseFolder, originalPath)),
	})
	return err
}
