package metadefender

import (
	"fmt"

	"icapeg/dtos"
	"icapeg/utils"
)

// toSubmitResponse transforms a metadefender scan response to generic sample response
func toSubmitResponse(sr *dtos.MetaDefenderScanFileResponse) *dtos.SubmitResponse {
	submitResp := &dtos.SubmitResponse{}
	if sr.DataID != "" {
		submitResp.SubmissionExists = true
	}
	submitResp.SubmissionID = sr.DataID
	submitResp.SubmissionSampleID = sr.DataID
	return submitResp
}

// toSampleInfo transforms a metadefender report response to generic sample info response
func toSampleInfo(mr *dtos.MetaDefenderReportResponse, fmi dtos.FileMetaInfo, failThreshold int) *dtos.SampleInfo {

	svrty := utils.SampleSeverityOk
	mtiScore := fmt.Sprintf("%d/%d", mr.ScanResults.TotalDetectedAvs, mr.ScanResults.TotalAvs)

	if mr.ScanResults.TotalDetectedAvs > failThreshold {
		svrty = utils.SampleSeverityMalicious
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

// toSubmissionStatusResponse transforms a metadefender report response to the generic submit status response
func toSubmissionStatusResponse(vr *dtos.MetaDefenderReportResponse) *dtos.SubmissionStatusResponse {
	submissionFinished := true
	if vr.ScanResults.ProgressPercentage < 100 {
		submissionFinished = false
	}

	return &dtos.SubmissionStatusResponse{
		SubmissionFinished: submissionFinished,
	}

}
