package transformers

import (
	"fmt"
	"icapeg/dtos"
	"icapeg/utils"
)

// the sample severity constants
const (
	GlasswallSampleSeverityOk        = "ok"
	GlasswallSampleSeverityMalicious = "malicious"
)

// TransformGlasswallToSubmitResponse transforms a Glasswall scan response to generic sample response
func TransformGlasswallToSubmitResponse(sr *dtos.GlasswallScanFileResponse) *dtos.SubmitResponse {
	submitResp := &dtos.SubmitResponse{}
	if sr.DataID != "" {
		submitResp.SubmissionExists = true
	}
	submitResp.SubmissionID = sr.DataID
	//submitResp.SubmissionID = sr.Resource // NOTE: this is done just for now, as virustotal doesn't make query with it's scan-id but rather it's resource id
	submitResp.SubmissionSampleID = sr.DataID
	return submitResp
}

// TransformGlasswallToSampleInfo transforms a Glasswall report response to generic sample info response
func TransformGlasswallToSampleInfo(mr *dtos.GlasswallReportResponse, fmi dtos.FileMetaInfo, failThreshold int) *dtos.SampleInfo {

	svrty := GlasswallSampleSeverityOk
	mtiScore := fmt.Sprintf("%d/%d", mr.ScanResults.TotalDetectedAvs, mr.ScanResults.TotalAvs)

	if mr.ScanResults.TotalDetectedAvs > failThreshold {
		svrty = GlasswallSampleSeverityMalicious
	}

	submissionFinished := true
	if mr.ScanResults.ProgressPercentage < 100 {
		submissionFinished = false
	}

	return &dtos.SampleInfo{
		SampleSeverity:     svrty,
		VTIScore:           mtiScore,
		FileName:           fmi.FileName,
		SampleType:         fmi.FileType,
		FileSizeStr:        fmt.Sprintf("%.2fmb", utils.ByteToMegaBytes(int(fmi.FileSize))),
		SubmissionFinished: submissionFinished,
	}
}

// TransformGlasswallToSubmissionStatusResponse transforms a virustotal report response to the generic submit status response
func TransformGlasswallToSubmissionStatusResponse(vr *dtos.GlasswallReportResponse) *dtos.SubmissionStatusResponse {
	submissionFinished := true
	if vr.ScanResults.ProgressPercentage < 100 {
		submissionFinished = false
	}

	return &dtos.SubmissionStatusResponse{
		SubmissionFinished: submissionFinished,
	}

}
