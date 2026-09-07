package models

import "io"

type HeadObjectRequest struct {
	Bucket string
	Key    string
}

type HeadObjectResponse struct {
	MimeType  *string
	SizeBytes *int64
}

type GetObjectRequest struct {
	Bucket string
	Key    string
	Range  string
}

type GetObjectResponse struct {
	Body         io.ReadCloser
	ContentRange *string
}

type DeleteObjectRequest struct {
	Bucket string
	Key    string
}
type DeleteObjectsRequest struct {
	Bucket string
	Keys   []string
}

type DeleteObjectsResponse struct {
	Deleted []string
}
