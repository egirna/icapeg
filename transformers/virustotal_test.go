package transformers

import (
	"icapeg/dtos"
	"testing"
)

func TestTransformVirusTotalToSubmitResponse(t *testing.T) {
	type testSample struct {
		sr *dtos.VirusTotalScanFileResponse
		submitResp *dtos.SubmitResponse
	}
	sampleTable := []testSample{
		{
			sr: &dtos.VirusTotalScanFileResponse{
				ResponseCode: 1,
			},
			submitResp: &dtos.SubmitResponse{
				SubmissionExists: true,
			},
		},
		{
			sr: &dtos.VirusTotalScanFileResponse{
				ResponseCode: 2,
			},
			submitResp: &dtos.SubmitResponse{
				SubmissionExists: false,
			},
		},
	}
	for _, sample := range sampleTable {
		got := TransformVirusTotalToSubmitResponse(sample.sr)
		want := sample.submitResp
		if got.SubmissionExists != want.SubmissionExists {
			t.Errorf("TransformMetaDefenderToSubmitResponse Failed for %s , wanted: %v got: %v",
				"SubmissionExists", want.SubmissionExists, got.SubmissionExists)
		}
	}
}
