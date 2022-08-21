package grayimages

import (
	"icapeg/utils"
	"log"
)

func (g *GrayImages) Processing(partial bool) (int, interface{}, map[string]string) {
	serviceHeaders := make(map[string]string)
	// no need to scan part of the file, this service needs all the file at ine time
	if partial {
		return utils.Continue, nil, nil
	}

	isGzip := false

	//extracting the file from http message
	file, reqContentType, err := g.generalFunc.CopyingFileToTheBuffer(g.methodName)
	if err != nil {
		log.Println("30")
		return utils.InternalServerErrStatusCodeStr, nil, serviceHeaders
	}

	//getting the extension of the file
	contentType := g.httpMsg.Response.Header["Content-Type"]
	var fileName string
	if g.methodName == utils.ICAPModeReq {
		fileName = utils.GetFileName(g.httpMsg.Request)
	} else {
		fileName = utils.GetFileName(g.httpMsg.Response)
	}
	fileExtension := utils.GetMimeExtension(file.Bytes(), contentType[0], fileName)

	isProcess, icapStatus, httpMsg := g.generalFunc.CheckTheExtension(fileExtension, g.extArrs,
		g.processExts, g.rejectExts, g.bypassExts, g.return400IfFileExtRejected, isGzip,
		g.serviceName, g.methodName, VirustotalIdentifier, g.httpMsg.Request.RequestURI, reqContentType, file)
	if !isProcess {
		return icapStatus, httpMsg, serviceHeaders
	}

}
