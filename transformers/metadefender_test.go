package transformers

import (
	"icapeg/dtos"
	"testing"
)

func TestTransformMetaDefenderToSubmitResponse(t *testing.T) {
	type testSample struct {
		sr *dtos.MetaDefenderScanFileResponse
		result *dtos.SubmitResponse
	}
	sampleTable := []testSample{
		{
			sr: &dtos.MetaDefenderScanFileResponse{
				DataID: "1",
			},
			result: &dtos.SubmitResponse{
				SubmissionID: "1",
				SubmissionSampleID: "1",
				SubmissionExists: true,
			},
		},
		{
			sr: &dtos.MetaDefenderScanFileResponse{
				DataID: "2",
			},
			result: &dtos.SubmitResponse{
				SubmissionID: "2",
				SubmissionSampleID: "2",
				SubmissionExists: true,
			},
		},
		{
			sr: &dtos.MetaDefenderScanFileResponse{
				DataID: "",
			},
			result: &dtos.SubmitResponse{
				SubmissionID: "",
				SubmissionSampleID: "",
				SubmissionExists: false,
			},
		},
	}
	for _, sample := range sampleTable {
		got := TransformMetaDefenderToSubmitResponse(sample.sr)
		want := sample.result

		if got != want {
			if sample.sr.DataID == "" && got.SubmissionExists{
				t.Errorf("TransformMetaDefenderToSubmitResponse Failed for %s , wanted: %v got: %v",
					"SubmissionExists", want.SubmissionExists, got.SubmissionExists)
			}
			if sample.sr.DataID != got.SubmissionID {
				t.Errorf("TransformMetaDefenderToSubmitResponse Failed for %s , wanted: %s got: %s",
					"SubmissionID", want.SubmissionID, got.SubmissionID)
			}
			if sample.sr.DataID != got.SubmissionSampleID{
				t.Errorf("TransformMetaDefenderToSubmitResponse Failed for %s , wanted: %s got: %s",
					"SubmissionSampleID", want.SubmissionSampleID, got.SubmissionSampleID)
			}
		}
	}
}
