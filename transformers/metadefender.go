package transformers

import (
	"fmt"
	"icapeg/dtos"
	"icapeg/helpers"
	"reflect"

	"github.com/spf13/viper"
)

// the sample severity constants
const (
	MetaDefenderSampleSeverityOk        = "ok"
	MetaDefenderSampleSeverityMalicious = "malicious"
)

// TransformMetaDefenderToSubmitResponse transforms a metadefender scan response to generic sample response
func TransformMetaDefenderToSubmitResponse(sr *dtos.MetaDefenderScanFileResponse) *dtos.SubmitResponse {
	submitResp := &dtos.SubmitResponse{}
	if sr.DataID != "" {
		submitResp.SubmissionExists = true
	}
	submitResp.SubmissionID = sr.DataID
	//submitResp.SubmissionID = sr.Resource // NOTE: this is done just for now, as virustotal doesn't make query with it's scan-id but rather it's resource id
	submitResp.SubmissionSampleID = sr.DataID
	return submitResp
}

// TransformMetaDefenderToSampleInfo transforms a metadefender report response to generic sample info reponse
func TransformMetaDefenderToSampleInfo(vr *dtos.MetaDefenderReportResponse, fmi dtos.FileMetaInfo) *dtos.SampleInfo {

	v := reflect.ValueOf(vr.ScanResults.ScanDetails)
	failThreshold := viper.GetInt("metadefender.fail_threshold")

	failCount := 0

	svrty := MetaDefenderSampleSeverityOk
	mtiScore := fmt.Sprintf("%d/%d", vr.ScanResults.TotalAvs-vr.ScanResults.TotalDetectedAvs, vr.ScanResults.TotalAvs)

	for i := 0; i < v.NumField(); i++ {
		scnr := v.Field(i).Interface().(dtos.MDScan)
		if scnr.ThreatFound != "" {
			failCount++
		}
		if failCount > failThreshold {
			svrty = MetaDefenderSampleSeverityMalicious
			break
		}
	}

	submissionFinished := true
	if vr.ScanResults.ProgressPercentage < 100 {
		submissionFinished = false
	}

	return &dtos.SampleInfo{
		SampleSeverity:     svrty,
		VTIScore:           mtiScore,
		FileName:           fmi.FileName,
		SampleType:         fmi.FileType,
		FileSizeStr:        fmt.Sprintf("%.2fmb", helpers.ByteToMegaBytes(int(fmi.FileSize))),
		SubmissionFinished: submissionFinished,
	}
}

// TransformMetaDefenderToSubmissionStatusResponse transforms a virustotal report response to the generic submit status response
func TransformMetaDefenderToSubmissionStatusResponse(vr *dtos.MetaDefenderReportResponse) *dtos.SubmissionStatusResponse {
	submissionFinished := true
	if vr.ScanResults.ProgressPercentage < 100 {
		submissionFinished = false
	}

	return &dtos.SubmissionStatusResponse{
		SubmissionFinished: submissionFinished,
	}

}
