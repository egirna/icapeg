package transformers

import (
	"icapeg/dtos"
	"testing"
)

func TestTransformMetaDefenderToSubmitResponse(t *testing.T) {
	type testSample struct {
		sr *dtos.MetaDefenderScanFileResponse
		submitResp *dtos.SubmitResponse
	}
	sampleTable := []testSample{
		{
			sr: &dtos.MetaDefenderScanFileResponse{
				DataID: "1",
			},
			submitResp: &dtos.SubmitResponse{
				SubmissionID: "1",
				SubmissionSampleID: "1",
				SubmissionExists: true,
			},
		},
		{
			sr: &dtos.MetaDefenderScanFileResponse{
				DataID: "2",
			},
			submitResp: &dtos.SubmitResponse{
				SubmissionID: "2",
				SubmissionSampleID: "2",
				SubmissionExists: true,
			},
		},
		{
			sr: &dtos.MetaDefenderScanFileResponse{
				DataID: "",
			},
			submitResp: &dtos.SubmitResponse{
				SubmissionID: "",
				SubmissionSampleID: "",
				SubmissionExists: false,
			},
		},
	}
	for _, sample := range sampleTable {
		got := TransformMetaDefenderToSubmitResponse(sample.sr)
		want := sample.submitResp

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

func TestTransformMetaDefenderToSubmissionStatusResponse(t *testing.T) {
	type testSample struct {
		vr *dtos.MetaDefenderReportResponse
		subStatResp *dtos.SubmissionStatusResponse
	}
	sampleTable := []testSample{
		{
			vr: &dtos.MetaDefenderReportResponse{
				ScanResults : dtos.MetaDefenderScanResults{
					ProgressPercentage: 50,
				},

			},
			subStatResp: &dtos.SubmissionStatusResponse{
				SubmissionFinished: false,
			},
		},
		{
			vr: &dtos.MetaDefenderReportResponse{
				ScanResults : dtos.MetaDefenderScanResults{
					ProgressPercentage: 100,
				},

			},
			subStatResp: &dtos.SubmissionStatusResponse{
				SubmissionFinished: true,
			},
		},
		{
			vr: &dtos.MetaDefenderReportResponse{
				ScanResults : dtos.MetaDefenderScanResults{
					ProgressPercentage: 150,
				},

			},
			subStatResp: &dtos.SubmissionStatusResponse{
				SubmissionFinished: true,
			},
		},
	}
	for _, sample := range sampleTable {
		got := TransformMetaDefenderToSubmissionStatusResponse(sample.vr)
		want := sample.subStatResp
		if got.SubmissionFinished != want.SubmissionFinished {
			t.Errorf("TransformMetaDefenderToSubmissionStatusResponse Failed for %s , wanted: %v got: %v",
				"SubmissionFinished", want.SubmissionFinished, got.SubmissionFinished)
		}
	}
}
