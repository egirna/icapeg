package glasswall

import (
	"fmt"

	"icapeg/dtos"
	"icapeg/utils"
)

// toSubmissionStatusResponse transforms a Glasswall scan response to generic sample response
func toSubmissionStatusResponse(reportResponse *dtos.GlasswallReportResponse) *dtos.SubmissionStatusResponse {
	submissionFinished := true
	if reportResponse.ScanResults.ProgressPercentage < 100 {
		submissionFinished = false
	}

	return &dtos.SubmissionStatusResponse{
		SubmissionFinished: submissionFinished,
	}
}

// toSubmitResponse transforms a Glasswall report response to generic sample info response
func toSubmitResponse(gwScanFileResponse *dtos.GlasswallScanFileResponse) *dtos.SubmitResponse {
	submitResp := &dtos.SubmitResponse{}
	if gwScanFileResponse.DataID != "" {
		submitResp.SubmissionExists = true
	}
	submitResp.SubmissionID = gwScanFileResponse.DataID
	submitResp.SubmissionSampleID = gwScanFileResponse.DataID
	return submitResp
}

// toSampleInfo transforms a Glasswall report response to the generic submit status response
func toSampleInfo(reportResponse *dtos.GlasswallReportResponse, fmi dtos.FileMetaInfo, failThreshold int) *dtos.SampleInfo {
	svrty := utils.SampleSeverityOk
	mtiScore := fmt.Sprintf("%d/%d", reportResponse.ScanResults.TotalDetectedAvs, reportResponse.ScanResults.TotalAvs)

	if reportResponse.ScanResults.TotalDetectedAvs > failThreshold {
		svrty = utils.SampleSeverityMalicious
	}

	submissionFinished := true
	if reportResponse.ScanResults.ProgressPercentage < 100 {
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
