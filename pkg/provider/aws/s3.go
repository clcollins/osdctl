package aws

import (
	"k8s.io/klog"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3"
)

// Delete all S3 buckets with the specified prefix
func DeleteS3BucketsWithPrefix(awsClient Client, prefix string) error {
	resp, err := awsClient.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return err
	}

	for _, bucket := range resp.Buckets {
		if strings.HasPrefix(*bucket.Name, prefix) {
			klog.Infoln("Deleting bucket", *bucket.Name)

			objects, err := awsClient.ListObjects(&s3.ListObjectsInput{
				Bucket: bucket.Name,
			})
			if err != nil {
				return err
			}

			// Clean up the objects in the bucket
			if len(objects.Contents) > 0 {
				deleteObjects := make([]*s3.ObjectIdentifier, 0, len(objects.Contents))
				for _, obj := range objects.Contents {
					deleteObjects = append(deleteObjects, &s3.ObjectIdentifier{Key: obj.Key})
				}

				if _, err = awsClient.DeleteObjects(
					&s3.DeleteObjectsInput{
						Delete: &s3.Delete{Objects: deleteObjects},
						Bucket: bucket.Name,
					},
				); err != nil {
					klog.Errorf("Failed to delete objects in bucket %s: %v", *bucket.Name, err)
					return err
				}
			}

			if _, err = awsClient.DeleteBucket(&s3.DeleteBucketInput{
				Bucket: bucket.Name}); err != nil {
				klog.Errorf("Failed to delete bucket %s: %v", *bucket.Name, err)
				return err
			}
		}
	}
	return nil
}
