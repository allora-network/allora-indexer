package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/schollz/progressbar/v3"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/rs/zerolog/log"
)

type progressReader struct {
	reader io.Reader
	bar    *progressbar.ProgressBar
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.bar.Add(n)
	return n, err
}

func restoreBackupFromS3() (string, error) {
	log.Info().Msg("Downloading SQL file from S3...")
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"), // Update with your region
		Credentials: credentials.NewStaticCredentials(awsAccessKey, awsSecretKey, ""),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create AWS session: %v", err)
	}

	s3Client := s3.New(sess)

	// Check if we need to fetch the latest file key
	if s3FileKey == "latest" {
		latestFileKey, err := getLatestFileKey(s3Client)
		if err != nil {
			return "", fmt.Errorf("failed to get latest file key: %v", err)
		}
		s3FileKey = latestFileKey
	}

	tempFile, err := os.CreateTemp("", "backup-*.dump")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer tempFile.Close()

	resp, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s3BucketName),
		Key:    aws.String(s3FileKey),
	})
	if err != nil {
		return "", fmt.Errorf("failed to download file from S3: %v", err)
	}
	defer resp.Body.Close()

	// Create a progress bar
	bar := progressbar.DefaultBytes(
		*resp.ContentLength,
		"Downloading",
	)

	// Create a progress reader
	progressReader := &progressReader{
		reader: resp.Body,
		bar:    bar,
	}

	// Copy the data from the progress reader to the temp file
	_, err = io.Copy(tempFile, progressReader)
	if err != nil {
		return "", fmt.Errorf("failed to read from S3 response body: %v", err)
	}

	fileName := tempFile.Name()

	_ = restoreBackupToDB(fileName)
	return tempFile.Name(), nil
}

func getLatestFileKey(s3Client *s3.S3) (string, error) {
	resp, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s3BucketName),
		Key:    aws.String("latest_backup.txt"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get latest_backup.txt: %v", err)
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read latest_backup.txt content: %v", err)
	}

	latestFileKey := strings.TrimSpace(string(content))
	if latestFileKey == "" {
		return "", fmt.Errorf("latest_backup.txt is empty")
	}

	return latestFileKey, nil
}

func gunzipFile(src string) error {
	cmd := exec.Command("gunzip", src)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run gunzip command: %v", err)
	}

	log.Info().Msg("DUMP file extracted.")
	return nil
}

func restoreBackupToDB(filePath string) error {

	cmd := exec.Command(
		"pg_restore",
		"-h", dbPool.Config().ConnConfig.Host,
		"-p", fmt.Sprintf("%d", dbPool.Config().ConnConfig.Port),
		"-U", dbPool.Config().ConnConfig.User,
		"-d", dbPool.Config().ConnConfig.Database,
		"-j", fmt.Sprintf("%d", parallelJobs),
		"-v", filePath,
	)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+dbPool.Config().ConnConfig.Password)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Err(err).Msg("Failed to run pg_restore command")
		return err
	}

	log.Info().Msg("Restored DB from DUMP")
	return nil
}
