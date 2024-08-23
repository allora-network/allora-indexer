package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/rs/zerolog/log"
)

func downloadBackupFromS3() (string, error) {
	log.Info().Msg("Downloading SQL file from S3...")
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"), // Update with your region
		Credentials: credentials.NewStaticCredentials(awsAccessKey, awsSecretKey, ""),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create AWS session: %v", err)
	}

	s3Client := s3.New(sess)

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

	_, err = tempFile.ReadFrom(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read from S3 response body: %v", err)
	}
	fileName := tempFile.Name()

	_ = restoreBackupToDB(fileName)
	return tempFile.Name(), nil
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
