package transformers

import (
	"fmt"
	"icapeg/dtos"
	"icapeg/utils"
)

// the sample severity constants
const (
	VirusTotalSampleSeverityOk        = "ok"
	VirusTotalSampleSeverityMalicious = "malicious"
)

// TransformVirusTotalToSubmitResponse transforms a virustotal scan response to generic sample response
func TransformVirusTotalToSubmitResponse(sr *dtos.VtUploadData) *dtos.SubmitResponse {
	submitResp := &dtos.SubmitResponse{}
	/*if sr.ResponseCode == 1 {
		submitResp.SubmissionExists = true
	}*/
	// submitResp.SubmissionID = sr.ScanID
	submitResp.SubmissionExists = true
	submitResp.SubmissionID = sr.Id // NOTE: this is done just for now, as virustotal doesn't make query with it's scan-id but rather it's resource id
	submitResp.SubmissionSampleID = sr.Id
	return submitResp
}

// TransformVirusTotalToSampleInfo transforms a virustotal report response to generic sample info response
func TransformVirusTotalToSampleInfo(vr *dtos.VirusTotalReportResponseV3, fmi dtos.FileMetaInfo, failThreshold int) *dtos.SampleInfo {

	svrty := VirusTotalSampleSeverityOk
	total := 73
	positives := vr.Data.Attributes.Stats.Suspicious + vr.Data.Attributes.Stats.Malicious + vr.Data.Attributes.Stats.Harmless
	vtiScore := fmt.Sprintf("%d/%d", positives, total)

	if int(positives) > failThreshold {
		svrty = VirusTotalSampleSeverityMalicious
	}
	submissionFinished := false
	if vr.Data.Attributes.Status == "completed" {
		//fmt.Println(vr.Data)
		submissionFinished = true
	}

	return &dtos.SampleInfo{
		SampleSeverity:     svrty,
		VTIScore:           vtiScore,
		FileName:           fmi.FileName,
		SampleType:         fmi.FileType,
		FileSizeStr:        fmt.Sprintf("%.2fmb", utils.ByteToMegaBytes(int(fmi.FileSize))),
		SubmissionFinished: submissionFinished,
	}

}

// TransformVirusTotalToSubmissionStatusResponse transforms a virustotal report response to the generic submit status response
func TransformVirusTotalToSubmissionStatusResponse(vr *dtos.VirusTotalReportResponse) *dtos.SubmissionStatusResponse {
	submissionFinished := true
	if vr.ResponseCode < 1 {
		submissionFinished = false
	}

	return &dtos.SubmissionStatusResponse{
		SubmissionFinished: submissionFinished,
	}

}
