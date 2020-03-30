package api

import (
	"bytes"
	"fmt"
	"icapeg/dtos"
	"icapeg/helpers"
	"icapeg/service"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/egirna/icap"

	"github.com/spf13/viper"
)

//ISTag exportable variable
var ISTag = "\"ICAPEG\""

// ToICAPEGResp is the ICAP Response Mode Handler:
func ToICAPEGResp(w icap.ResponseWriter, req *icap.Request) {
	h := w.Header()
	h.Set("ISTag", ISTag)
	h.Set("Service", "Egirna ICAP-EG")

	log.Printf("Request received---> METHOD:%s URL:%s\n", req.Method, req.RawURL)

	switch req.Method {
	case "OPTIONS":
		h.Set("Methods", "RESPMOD")
		h.Set("Allow", "204")
		h.Set("Preview", viper.GetString("app.preview_bytes"))
		h.Set("Transfer-Preview", "*")
		w.WriteHeader(http.StatusOK, nil, false)
	case "RESPMOD":

		// getting the content type to determine if the response is for a file, and if so, if its allowed to be processed
		// according to the configuration

		ct := helpers.GetMimeExtension(req.Preview)

		if helpers.InStringSlice(ct, viper.GetStringSlice("app.unprocessable_extensions")) {
			log.Println("Processing not required")
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}

		// copying the file to a buffer for scanner vendor processing as the file is allowed according the co figuration

		buf := &bytes.Buffer{}

		if _, err := io.Copy(buf, req.Response.Body); err != nil {
			log.Println("Failed to copy the response body to buffer: ", err.Error())
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}

		if buf.Len() > viper.GetInt("app.max_filesize") {
			log.Println("File size too large")
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}

		// preparing the file meta informations
		filename := helpers.GetFileName(req.Request)
		fileExt := helpers.GetFileExtension(req.Request)
		fmi := dtos.FileMetaInfo{
			FileName: filename,
			FileType: fileExt,
			FileSize: float64(buf.Len()),
		}

		if viper.GetBool("app.local_scanner") {
			lsvc := service.GetLocalService(strings.ToLower(viper.GetString("app.scanner_vendor")))

			if lsvc == nil {
				log.Println("No such scanner vendors:", viper.GetString("app.scanner_vendor"))
				w.WriteHeader(helpers.IfPropagateError(http.StatusBadRequest, http.StatusNoContent), nil, false)
				return
			}

			sampleInfo, err := lsvc.ScanFileStream(buf, fmi)
			if err != nil {
				log.Println("Couldn't fetch sample information for local service: ", err.Error())
				w.WriteHeader(helpers.IfPropagateError(http.StatusFailedDependency, http.StatusNoContent), nil, false)
				return
			}

			if !helpers.InStringSlice(sampleInfo.SampleSeverity, viper.GetStringSlice(helpers.GetScannerVendorSpecificCfg("ok_file_status"))) { // checking is the sample severity is amongst the allowable file status
				log.Printf("The file:%s is %s\n", filename, sampleInfo.SampleSeverity)
				htmlBuf, newResp := helpers.GetTemplateBufferAndResponse(helpers.BadFileTemplate, &dtos.TemplateData{
					FileName:     sampleInfo.FileName,
					FileType:     sampleInfo.SampleType,
					FileSizeStr:  sampleInfo.FileSizeStr,
					RequestedURL: "N/A",
					Severity:     sampleInfo.SampleSeverity,
					Score:        sampleInfo.VTIScore,
					ResultsBy:    viper.GetString("app.scanner_vendor"),
				})
				w.WriteHeader(http.StatusOK, newResp, true)
				w.Write(htmlBuf.Bytes())

				return

			}

			log.Printf("The file %s is good to go\n", sampleInfo.FileName)
			w.WriteHeader(http.StatusNoContent, nil, false) // all ok, show the contents as it is
			return

		}

		// making necessary service api calls

		svc := service.GetService(strings.ToLower(viper.GetString("app.scanner_vendor")))

		if svc == nil {
			log.Println("No such scanner vendors:", viper.GetString("app.scanner_vendor"))
			w.WriteHeader(helpers.IfPropagateError(http.StatusBadRequest, http.StatusNoContent), nil, false)
			return
		}

		// The submit file api call is commented out for safety for now
		submitResp, err := svc.SubmitFile(buf, filename) // submitting the file for analysing
		if err != nil {
			log.Printf("Failed to submit file to %s: %s\n", viper.GetString("app.scanner_vendor"), err.Error())
			w.WriteHeader(helpers.IfPropagateError(http.StatusFailedDependency, http.StatusNoContent), nil, false)
			return
		}

		if !submitResp.SubmissionExists {
			log.Println("No submissions for the file")
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}

		if viper.GetBool("app.debug") {
			spew.Dump("submit response", submitResp)
		}

		submissionFinished := false
		statusCheckFinishTime := time.Now().Add(viper.GetDuration(helpers.GetScannerVendorSpecificCfg("status_check_timeout")) * time.Second) // the time after which, the system is to stop checking for submission finish
		var sampleInfo *dtos.SampleInfo
		sampleID := submitResp.SubmissionSampleID //"4715575"

		for !submissionFinished && time.Now().Before(statusCheckFinishTime) {
			submissionID := submitResp.SubmissionID //"5651578"

			switch viper.GetBool(helpers.GetScannerVendorSpecificCfg("status_endpoint_exists")) {
			case true:
				submissionStatus, err := svc.GetSubmissionStatus(submissionID) // getting the file submission status by the submission id received by submitting the file
				if err != nil {
					log.Printf("Failed to get submission status from %s: %s\n", viper.GetString("app.scanner_vendor"), err.Error())
					w.WriteHeader(helpers.IfPropagateError(http.StatusFailedDependency, http.StatusNoContent), nil, false)
					return
				}

				if viper.GetBool("app.debug") {
					spew.Dump("submission status resp", submissionStatus)
				}
				submissionFinished = submissionStatus.SubmissionFinished
			case false:
				var err error
				sampleInfo, err = svc.GetSampleFileInfo(sampleID, fmi)
				if err != nil {
					log.Println("Couldn't fetch sample information during status check: ", err.Error())
					w.WriteHeader(helpers.IfPropagateError(http.StatusFailedDependency, http.StatusNoContent), nil, false)
					return
				}
				submissionFinished = sampleInfo.SubmissionFinished
			default:
				log.Println("Put the status_endpoint_exists field in the config file under the scanner vendor")
			}

			if !submissionFinished { // if the submission is not finished, wait for a certain time and then call again
				time.Sleep(viper.GetDuration(helpers.GetScannerVendorSpecificCfg("status_check_interval")) * time.Second)
			}
		}

		if submissionFinished {
			if sampleInfo == nil {
				var err error
				sampleInfo, err = svc.GetSampleFileInfo(sampleID, fmi) // getting the results after scanner is done analysing the file
				if err != nil {
					log.Println("Couldn't fetch sample information after submission finish: ", err.Error())
					w.WriteHeader(helpers.IfPropagateError(http.StatusFailedDependency, http.StatusNoContent), nil, false)
					return
				}
			}

			if !helpers.InStringSlice(sampleInfo.SampleSeverity, viper.GetStringSlice(helpers.GetScannerVendorSpecificCfg("ok_file_status"))) { // checking is the sample severity is amongst the allowable file status
				log.Printf("The file:%s is %s\n", filename, sampleInfo.SampleSeverity)
				htmlBuf, newResp := helpers.GetTemplateBufferAndResponse(helpers.BadFileTemplate, &dtos.TemplateData{
					FileName:     sampleInfo.FileName,
					FileType:     sampleInfo.SampleType,
					FileSizeStr:  sampleInfo.FileSizeStr,
					RequestedURL: helpers.BreakHTTPURL(req.Request.RequestURI),
					Severity:     sampleInfo.SampleSeverity,
					Score:        sampleInfo.VTIScore,
					ResultsBy:    viper.GetString("app.scanner_vendor"),
				})
				w.WriteHeader(http.StatusOK, newResp, true)
				w.Write(htmlBuf.Bytes())

				return
			}
		} else { // this can only mean that the current time has crossed the statusCheckFinishTime
			log.Println("File submission is taking too long to finish")
			w.WriteHeader(http.StatusNoContent, nil, false)
			return
		}

		log.Printf("The file %s is good to go\n", sampleInfo.FileName)
		w.WriteHeader(http.StatusNoContent, nil, false) // all ok, show the contents as it is
	case "ERRDUMMY":
		w.WriteHeader(http.StatusBadRequest, nil, false)
		fmt.Println("Malformed request")
	default:
		w.WriteHeader(http.StatusMethodNotAllowed, nil, false)
		fmt.Println("Invalid request method - respmod")
	}
}
