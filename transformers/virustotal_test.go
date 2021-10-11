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

func TestTransformVirusTotalToSampleInfo(t *testing.T) {
	type testSample struct {
		vr *dtos.VirusTotalReportResponse
		fmi dtos.FileMetaInfo
		failThreshold int
		result *dtos.SampleInfo
	}
	sampleTable := []testSample{
		{
			vr: &dtos.VirusTotalReportResponse{
				Positives: 50,
				ResponseCode: 4,
			},
			failThreshold: 30,
			result: &dtos.SampleInfo{
				SampleSeverity: VirusTotalSampleSeverityMalicious,
				SubmissionFinished: true,
			},
		},
		{
			vr: &dtos.VirusTotalReportResponse{
				Positives: 50,
				ResponseCode: 4,
			},
			failThreshold: 70,
			result: &dtos.SampleInfo{
				SampleSeverity:  MetaDefenderSampleSeverityOk,
				SubmissionFinished: true,
			},
		},
		{
			vr: &dtos.VirusTotalReportResponse{
				Positives: 50,
				ResponseCode: 0,
			},
			failThreshold: 70,
			result: &dtos.SampleInfo{
				SampleSeverity: MetaDefenderSampleSeverityOk,
				SubmissionFinished: false,
			},
		},
	}
	for _, sample := range sampleTable {
		got := TransformVirusTotalToSampleInfo(sample.vr, sample.fmi, sample.failThreshold)
		want := sample.result
		if got.SubmissionFinished != want.SubmissionFinished {
			t.Errorf("TransformMetaDefenderToSampleInfo Failed for %s , wanted: %v got: %v",
				"SubmissionFinished", want.SubmissionFinished, got.SubmissionFinished)
		}
		if got.SampleSeverity != want.SampleSeverity {
			t.Errorf("TransformMetaDefenderToSampleInfo Failed for %s , wanted: %v got: %v",
				"SampleSeverity", want.SampleSeverity, got.SampleSeverity)
		}
	}
}

