package transformers

import (
	"icapeg/dtos"
	"strconv"
	"testing"
)

//unit test
func TestTransformVmrayToSubmitResponse(t *testing.T) {
	type testSample struct {
		vsr *dtos.VmraySubmitResponse
		submitResp *dtos.SubmitResponse
	}
	sampleTable := []testSample{
		{
			vsr: &dtos.VmraySubmitResponse{
				Data: dtos.VmraySubmitData{
					Submissions: []dtos.VmraySubmissions{dtos.VmraySubmissions{
						SubmissionID: 1,
						SubmissionSampleID: 2,
					}, dtos.VmraySubmissions{
						SubmissionID: 3,
						SubmissionSampleID: 4,
					},dtos.VmraySubmissions{
						SubmissionID: 5,
						SubmissionSampleID: 6,
					}},
				},
			},
			submitResp: &dtos.SubmitResponse{
				SubmissionID: strconv.Itoa(1),
				SubmissionSampleID: strconv.Itoa(2),
				SubmissionExists: true,
			},
		},
		{
			vsr: &dtos.VmraySubmitResponse{
				Data: dtos.VmraySubmitData{
					Submissions: []dtos.VmraySubmissions{},
				},
			},
			submitResp: &dtos.SubmitResponse{
				SubmissionExists: false,
			},
		},
	}
	for _, sample := range sampleTable {
		got := TransformVmrayToSubmitResponse(sample.vsr)
		want := sample.submitResp
		if got.SubmissionExists != want.SubmissionExists {

		}
		if got.SubmissionExists != want.SubmissionExists {
			t.Errorf("TransformVirusTotalToSubmitResponse Failed for %s , wanted: %v got: %v",
				"SubmissionExists", want.SubmissionExists, got.SubmissionExists)
		}
		if got.SubmissionExists == want.SubmissionExists && got.SubmissionID != want.SubmissionID {
			t.Errorf("TransformVirusTotalToSampleInfo Failed for %s , wanted: %s got: %s",
				"SubmissionID", want.SubmissionID, got.SubmissionID)
		}
		if got.SubmissionExists == want.SubmissionExists && got.SubmissionSampleID != want.SubmissionSampleID {
			t.Errorf("TransformVirusTotalToSampleInfo Failed for %s , wanted: %s got: %s",
				"SubmissionExists", want.SubmissionSampleID, got.SubmissionSampleID)
		}
	}
}
