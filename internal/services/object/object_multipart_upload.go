package object

import (
	"context"
	"io"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// multipartUploadThreshold is the minimum object size, in bytes, for which we
// switch from a single PutObject to a concurrent multipart upload.
const multipartUploadThreshold int64 = 5 * 1024 * 1024 // 5 MiB

// multipartUploadPartSize is the size of each uploaded part. Must be at least
// 5 MiB for every part except the trailing one.
const multipartUploadPartSize int64 = 16 * 1024 * 1024 // 16 MiB

// multipartUploadConcurrency is the number of parts uploaded in parallel.
const multipartUploadConcurrency = 4

// uploadObject uploads the body in req.
func uploadObject(ctx context.Context, s3Client *s3.Client, req *s3.PutObjectInput) error {
	file, ok := req.Body.(*os.File)
	if !ok {
		_, err := s3Client.PutObject(ctx, req)

		return err
	}

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	if stat.Size() < multipartUploadThreshold {
		_, err := s3Client.PutObject(ctx, req)

		return err
	}

	return uploadObjectMultipart(ctx, s3Client, req, file, stat.Size())
}

// uploadObjectMultipart splits the file into fixed-size parts and uploads them
// concurrently.
func uploadObjectMultipart(
	ctx context.Context, s3Client *s3.Client, req *s3.PutObjectInput, file *os.File, size int64,
) error {
	initResp, err := s3Client.CreateMultipartUpload(ctx, &s3.CreateMultipartUploadInput{
		Bucket:               req.Bucket,
		Key:                  req.Key,
		ACL:                  req.ACL,
		ContentType:          req.ContentType,
		Metadata:             req.Metadata,
		StorageClass:         req.StorageClass,
		SSECustomerAlgorithm: req.SSECustomerAlgorithm,
		SSECustomerKey:       req.SSECustomerKey,
		SSECustomerKeyMD5:    req.SSECustomerKeyMD5,
	})
	if err != nil {
		return err
	}

	uploadID := initResp.UploadId

	// Abort unless complete succeeds. A no-op if the upload was already aborted
	// by the server. We ignore the error in order to return the original error.
	abort := func() {
		_, _ = s3Client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
			Bucket:   req.Bucket,
			Key:      req.Key,
			UploadId: uploadID,
		})
	}

	partCount := int((size + multipartUploadPartSize - 1) / multipartUploadPartSize)
	parts := make([]s3Types.CompletedPart, partCount)

	type partJob struct {
		number int32
		offset int64
	}

	jobs := make(chan partJob, partCount)

	var (
		wg       sync.WaitGroup
		firstErr error
		errOnce  sync.Once
	)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	worker := func() {
		for j := range jobs {
			if ctx.Err() != nil {
				return
			}

			section := io.NewSectionReader(file, j.offset, multipartUploadPartSize)
			partNum := j.number
			partLen := section.Size()

			if j.offset+partLen > size {
				partLen = size - j.offset
			}

			resp, err := s3Client.UploadPart(ctx, &s3.UploadPartInput{
				Bucket:               req.Bucket,
				Key:                  req.Key,
				UploadId:             uploadID,
				PartNumber:           &partNum,
				Body:                 section,
				ContentLength:        &partLen,
				SSECustomerAlgorithm: req.SSECustomerAlgorithm,
				SSECustomerKey:       req.SSECustomerKey,
				SSECustomerKeyMD5:    req.SSECustomerKeyMD5,
			})
			if err != nil {
				errOnce.Do(func() {
					firstErr = err

					cancel()
				})

				return
			}

			parts[j.number-1] = s3Types.CompletedPart{
				ETag:       resp.ETag,
				PartNumber: aws.Int32(j.number),
			}
		}
	}

	for range multipartUploadConcurrency {
		wg.Go(worker)
	}

	for i := range partCount {
		jobs <- partJob{
			number: int32(i + 1),
			offset: int64(i) * multipartUploadPartSize,
		}
	}

	close(jobs)
	wg.Wait()

	if firstErr != nil {
		abort()

		return firstErr
	}

	_, err = s3Client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:          req.Bucket,
		Key:             req.Key,
		UploadId:        uploadID,
		MultipartUpload: &s3Types.CompletedMultipartUpload{Parts: parts},
	})
	if err != nil {
		abort()

		return err
	}

	return nil
}
