package transformers

import (
	"fmt"
	"icapeg/dtos"
	"icapeg/utils"
	"reflect"

	"github.com/spf13/viper"
)

// the sample severity constants
const (
	VirusTotalSampleSeverityOk        = "ok"
	VirusTotalSampleSeverityMalicious = "malicious"
)

// TransformVirusTotalToSubmitResponse transforms a virustotal scan response to generic sample response
func TransformVirusTotalToSubmitResponse(sr *dtos.VirusTotalScanFileResponse) *dtos.SubmitResponse {
	submitResp := &dtos.SubmitResponse{}
	if sr.ResponseCode == 1 {
		submitResp.SubmissionExists = true
	}
	// submitResp.SubmissionID = sr.ScanID
	submitResp.SubmissionID = sr.Resource // NOTE: this is done just for now, as virustotal doesn't make query with it's scan-id but rather it's resource id
	submitResp.SubmissionSampleID = sr.Resource
	return submitResp
}

// TransformVirusTotalToSampleInfo transforms a virustotal report response to generic sample info response
func TransformVirusTotalToSampleInfo(vr *dtos.VirusTotalReportResponse, fmi dtos.FileMetaInfo) *dtos.SampleInfo {

	v := reflect.ValueOf(vr.Scans)

	failThreshold := viper.GetInt("virustotal.fail_threshold")

	failCount := 0

	svrty := VirusTotalSampleSeverityOk
	vtiScore := fmt.Sprintf("%d/%d", vr.Positives, vr.Total)

	for i := 0; i < v.NumField(); i++ {
		scnr := v.Field(i).Interface().(dtos.Scanner)
		if scnr.Detected {
			failCount++
		}
		if failCount > failThreshold {
			svrty = VirusTotalSampleSeverityMalicious
			break
		}
	}

	submissionFinished := true
	if vr.ResponseCode < 1 {
		submissionFinished = false
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
