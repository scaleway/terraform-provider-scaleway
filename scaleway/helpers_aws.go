package scaleway

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type ListObjectVersionsPaginator struct {
	params    *s3.ListObjectVersionsInput
	client    *s3.Client
	nextToken *string
	firstPage bool
}

func NewListObjectVersionsPaginator(params *s3.ListObjectVersionsInput) *ListObjectVersionsPaginator {
	if params == nil {
		params = &s3.ListObjectVersionsInput{}
	}

	return &ListObjectVersionsPaginator{
		params:    params,
		firstPage: true,
		nextToken: nil,
	}
}

func (p *ListObjectVersionsPaginator) HasMorePages() bool {
	return p.firstPage || p.nextToken != nil
}

func (p *ListObjectVersionsPaginator) NextPage(ctx context.Context) (*s3.ListObjectVersionsOutput, error) {
	if !p.HasMorePages() {
		return nil, fmt.Errorf("no more pages available")
	}

	params := *p.params
	params.VersionIdMarker = p.nextToken

	results, err := p.client.ListObjectVersions(ctx, &params)
	if err != nil {
		return nil, err
	}
	p.firstPage = false

	p.nextToken = nil
	if results.IsTruncated {
		p.nextToken = results.NextVersionIdMarker
	}

	return results, nil
}
