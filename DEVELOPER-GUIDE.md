<h1 align="center"><b>Developer Guide</b></h1>

## Table of Contents

- [Introduction](#introduction)
- [Functions](#functions)
  - [**CopyingFileToTheBuffer**](#copyingfiletothebuffer)
    - [**Description**](#description)
    - [**Parameters**](#paramaters)
    - [**Return Values**](#return-values)
  - [**CheckTheExtension**](#checktheextension)
    - [**Description**](#description-1)
    - [**Parameters**](#paramaters-1)
    - [**Return Values**](#return-values-1)
  - [**IsBodyGzipCompressed**](#isbodygzipcompressed)
    - [**Description**](#description-2)
    - [**Parameters**](#paramaters-2)
    - [**Return Values**](#return-values-2)
  - [**DecompressGzipBody**](#decompressgzipbody)
    - [**Description**](#description-3)
    - [**Parameters**](#paramaters-3)
    - [**Return Values**](#return-values-3)
  - [**IfMaxFileSizeExc**](#ifmaxfilesizeexc)
    - [**Description**](#description-4)
    - [**Parameters**](#paramaters-4)
    - [**Return Values**](#return-values-4)
  - [**IfMaxFileSizeExc**](#ifmaxfilesizeexc)
    - [**Description**](#description-5)
    - [**Parameters**](#paramaters-5)
    - [**Return Values**](#return-values-5)
  - [**GetFileName**](#getfilename)
    - [**Description**](#description-6)
    - [**Parameters**](#paramaters-6)
    - [**Return Values**](#return-values-6)
  - [**CompressFileGzip**](#compressfilegzip)
    - [**Description**](#description-7)
    - [**Parameters**](#paramaters-7)
    - [**Return Values**](#return-values-7)
  - [**ErrPageResp**](#errpageresp)
    - [**Description**](#description-8)
    - [**Parameters**](#paramaters-8)
    - [**Return Values**](#return-values-8)
  - [**GenHtmlPage**](#gethtmlpage)
    - [**Description**](#description-9)
    - [**Parameters**](#paramaters-9)
    - [**Return Values**](#return-values-9)
  - [**PreparingFileAfterScanning**](#preparingfileafterscanning)
    - [**Description**](#description-10)
    - [**Parameters**](#paramaters-10)
    - [**Return Values**](#return-values-10)
  - [**IfStatusIs204WithFile**](#ifstatusis204withfile)
    - [**Description**](#description-11)
    - [**Parameters**](#paramaters-11)
    - [**Return Values**](#return-values-11)
  - [**ReturningHttpMessageWithFile**](#returninghttpmessagewithfile)
    - [**Description**](#description-12)
    - [**Parameters**](#paramaters-12)
    - [**Return Values**](#return-values-12)


## Introduction

[**GeneralFunc**](./service/services-utilities/general-functions/general-functions.go) is a struct used to help developer who implement a new service of a new vendor. It has a lot of functions that the developer may need, they help him to do some tasks. These tasks are explained in that file.

```go
type GeneralFunc struct {
	httpMsg *http_message.HttpMsg
    xICAPMetadata string
}
```

So the developer can add [**GeneralFunc**](./service/services-utilities/general-functions/general-functions.go) field in the struct of the service which he implements, like [**Echo service**](service/services/echo/config.go), a field from type [**GeneralFunc**](./service/services-utilities/general-functions/general-functions.go) is defined as the last field in [**Echo service**](service/services/echo/config.go) struct as shown below.

```go
type Echo struct {
    xICAPMetadata string
	httpMsg                    *http_message.HttpMsg
	elapsed                    time.Duration
	serviceName                string
	methodName                 string
	maxFileSize                int
	bypassExts                 []string
	processExts                []string
	rejectExts                 []string
	extArrs                    []services_utilities.Extension
	returnOrigIfMaxSizeExc     bool
	return400IfFileExtRejected bool
	generalFunc                *general_functions.GeneralFunc //GeneralFunc field
}
```

First, let's discuss about the only field in [**GeneralFunc**](./service/services-utilities/general-functions/general-functions.go) which is an instance from type [**HttpMsg**](http-message/httpMessage.go).

```go
type HttpMsg struct {
	Request  *http.Request
	Response *http.Response
}
```

[**HttpMsg**](consts/httpMessage.go) is a struct which contains two fields. The first field is **HTTP reauest** and the second is **HTTP response**. It's used to encapsulate the HTTP request and response with each other because there are two modes in **ICAP**, **REQMOD** and **RESPMOD**.

[**HttpMsg**](consts/httpMessage.go) field in [**GeneralFunc**](./service/services-utilities/general-functions/general-functions.go) struct facilitates [**GeneralFunc**](./service/services-utilities/general-functions/general-functions.go) dealing with HTTP request and response.

**xICAPMetadata** is a string value of the ID of **ICAP** request sent to **ICAPeg**. 

## Functions

- ### **CopyingFileToTheBuffer**

  - #### **Description**

    It extracts the body of the **HTTP message** and get the content type of the body if it exists.

  - #### **Parameters**

    - **methodName** (string): possible values are (**REQMOD** and **RESPMOD**).

  - **Return Values**

     - Pointer from **bytes.Buffer** type which points to the body of the **HTTP message**.
     - Instance from [**ContentType**](service/services-utilities/ContentTypes/contentType.go) type which is used as indicator to the content type of the body in case of **REQMOD**, but in case of **RESPMOD** it will be nil. This value is important when we prepare the **HTTP request** to return it to the client.
     - Instance from **error** type to indicate if there is an error or not, the value will be nil if there is no error.

- ### **CheckTheExtension**

  - #### **Description**

    It extracts the body of the **HTTP message** and gets the content type of the body if it exists.

  - #### **Parameters**

    - **fileExtension** (string): The extension of the file.
    - **extArrs** ([[]services_utilities.Extension](service/services-utilities/utils.go)): Array of [**Extension**](service/services-utilities/utils.go) type stores all information about extensions arrays (bypass, reject and process).
    - **processExts** ([]string): Array of extensions which wanted to be processed.
    - **rejectExts** ([]string): Array of extensions which wanted to be rejected.
    - **bypassExts** ([]string): Array of extensions which wanted to be bypassed.
    - **return400IfFileExtRejected** (bool): Bool variable indicates if the service configuration aims to return **400** as an ICAP response status in case the current file extension is rejected.
    - **isGzip** (bool): Boolean variable indicates to if the **HTTP message** body was compressed in the **HTTP message** before starting processing.
    - **serviceName** (string): The name of the service which you implement.
    - **methodName** (string): possible values are (**REQMOD** and **RESPMOD**).
    - **identifier** (string): The identifier of the service (service API identifier for example).
    - **requestURI** (string): The requested URL that may cause getting a block page in case the file extension is rejected.
    - **reqContentType** ([ContentType](service/services-utilities/ContentTypes/contentType.go)): the content type of the **HTTP message** body before processing (multipart for example).
    - **file** (*bytes.Buffer): The body of **HTTP message**.

  - #### **Return Values**

     - **Bool** variable, possible values are **true** if the file extension exists in process extension.
     - Integer variable which is **ICAP** status should be returned from [**Processing**](service/service.go) function. It equals zero if the file extension exists in the process extensions array.
     - Instance from **error** type to indicate if there is an error or not, the value will be nil if there is no error.

- ### **IsBodyGzipCompressed**

  - #### **Description**

    It checks if the body of an **HTTP message** is compressed in **Gzip**. 

  - #### **Parameters**

    - **methodName** (string): possible values are (**REQMOD** and **RESPMOD**).

  - #### **Return Values**

    - **Bool** variable, possible values are **true** if the body is compressed in **Gzip** and false if the body is not.

- ### **ReqModErrPage**

  - #### **Description**

    It prepares and returns error page specified to the service included in http request to replace the original HTTP request with the new one (HTTP request with error page). 

  - #### **Parameters**

    - **reason** (string): The reason of returning this HTML error page and this reason will be typed in the page.
    - **serviceName** (string): The service's name which uses this function and this service name will be typed in the page.
    - **serviceName** (string): The service's identifier which uses this function and this service identifier will be typed in the page.

  - #### **Return Values**

    - Instance from ***bytes.Buffer** type which is the HTML page converted to bytes.Buffer type to be included later inside the new HTTP request.
    - Instance from ***http.Request** type which is the new HTTP request of the error page.
    - Instance from **error** type to indicate whether there is an error or not, the value will be nil if there is no error.

- ### **DecompressGzipBody**

  - #### **Description**

    It decompresses the file is compressed in **Gzip**.

  - #### **Parameters**

    - **file** (*bytes.Buffer): The body of **HTTP message**.

    #### **Return Values**

    - Instance from **bytes.Buffer** type which is the file after decompression.
    - Instance from **error** type to indicate if there is an error or not, the value will be nil if there is no error.

- ### **IfMaxFileSizeExc**

  - #### **Description**

    It takes the actions which should be taken if the body of the **HTTP message**.

  - #### **Parameters**

    - **returnOrigIfMaxSizeExc** (bool): The bool variable indicates if the service configuration aims to return the Original file in case the **HTTP message** body exceeds the max file size.
    - **serviceName** (string): The name of the service which you implement.
    - **file** (*bytes.Buffer): The body of **HTTP message**.

  - #### **Return Values**

    - Integer variable which is **ICAP** status should be returned from [**Processing**](service/service.go) function.
    - Pointer from **bytes.Buffer** which is a body of the **HTTP message** should be returned from [**Processing**](service/service.go) function.
    - **interface{}** which is the **HTTP message** which should be returned from [**Processing**](service/service.go) function.

- ### **GetFileName**

  - #### **Description**

    It gets the file name from the **HTTP message**.

  - #### **Parameters**

    No parameters.

  - #### **Return Values**

    - String value stores the name of the file. if the file name doesn't exist in the requested **URL** of the **HTTP message**, the file name will be **unnamed_file**.

- ### **CompressFileGzip**

  - #### **Description**

    It's used to compress the body of the **HTTP message** body.

  - #### **Parameters**

    - **scannedFile** ([]byte): Array of bytes which contains the file which you want to compress.

  - #### **Return Values**

    - **[]byte** array of bytes contains the body of **HTTP message** after compression.
    - Instance from **error** type to indicate whether there is an error or not, the value will be nil if there is no error.

- ### **ErrPageResp**

  - #### **Description**

    It prepares headers and status code of **HTTP response** which will include a block page as **HTML** file.

  - #### **Parameters**

    - **status** (int): The status of the **HTTP response**.
    - **pageContentLength** (int): The content length of the **HTML** page.

  - #### **Return Values**

    - Pointer to **http.Response** type.

- ### **GenHtmlPage**

  - #### **Description**

    It returns a block page as an **HTML** page.

  - #### **Parameters**

    - **path** (string): [The path of the block page](block-page.html).
    - **reason** (string): The reason for getting a block page.
    - **serviceName** (string): The name of the service which you implement.
    - **identifierId** (string): The identifier of the service (service API identifier for example).
    - **reqUrl** (string): The requested URL that causes getting a block page.

  - #### **Return Values**

    - Pointer to **bytes.Buffer** type which stores the block page.

- ### **PreparingFileAfterScanning**

  - #### **Description**

    It's used for preparing the HTTP response before returning it. It's used for **REQMOD** only.

  - #### **Parameters**

    - **scannedFile** ([]byte): Array of bytes which contains the file which you want to prepare.
    - **reqContentType** ([ContentType](service/services-utilities/ContentTypes/contentType.go)): the content type of the **HTTP message** body before processing (multipart for example).
    - **methodName** (string): possible values are (**REQMOD** and **RESPMOD**).

  - #### **Return Values**

    - **[]byte** array of bytes stores the file after scanning.

    

- ### **IfStatusIs204WithFile**

  - #### **Description**

    it handles the **HTTP message** if the status should be 204 no modifications.

  - #### **Parameters**

    - **methodName** (string): Possible values are (**REQMOD** and **RESPMOD**).
    - **status** (int): The status of the **HTTP response**.
    - **file** (*bytes.Buffer): The body of **HTTP message**.
    - **isGzip** (bool): Boolean variable indicates to if the **HTTP message** body was compressed in the **HTTP message** before starting processing.
    - **reqContentType** ([ContentType](service/services-utilities/ContentTypes/contentType.go)): the content type of the **HTTP message** body before processing (multipart for example).
    - **httpMessage** interface{}: The **HTTP message** which should be returned from [**Processing**](service/service.go) function.

  - #### **Return Values**

    - **[]byte** array of bytes stores the file after the function finishes handling **HTTP message**.
    - **interface{}** which is the **HTTP message** which should be returned from [**Processing**](service/service.go) function after the function finishes handling **HTTP message**.

- ### **ReturningHttpMessageWithFile**

  - #### **Description**

    It returns the suitable **HTTP message** (HTTP request, HTTP response).

  - #### **Parameters**

    - **methodName** (string): possible values are (**REQMOD** and **RESPMOD**).
    - **file** (*bytes.Buffer): The file which wanted to be included in **HTTP message**.

  - #### **Return Values**

    - **interface{}** which is the **HTTP message** which should be returned from [**Processing**](service/service.go) function.

- **IfICAPStatusIs204**

  - **Description**

    It prepares the HTTP message if the ICAP status code is 204.

  - #### **Parameters**

    - **methodName** (string): possible values are (**REQMOD** and **RESPMOD**).
    - **status** (int): The status of the **ICAP response**.
    - **file** (*bytes.Buffer): The body of **HTTP message**.
    - **isGzip** (bool): Boolean variable indicates to if the **HTTP message** body was compressed in the **HTTP message** before starting processing.
    - **reqContentType** ([ContentType](service/services-utilities/ContentTypes/contentType.go)): the content type of the **HTTP message** body before processing (multipart for example).
    - **httpMessage** interface{}: The **HTTP message** which should be returned from [**Processing**](service/service.go) function.

  - #### **Return Values**

    - **[]byte** which is array of bytes contains the body of the HTTP message.
    - **interface{}** which is the **HTTP message** which should be returned from [**Processing**](service/service.go) function.

- **GetDecodedImage**

  - **Description**

    It takes the HTTP file and converts it to an image object.

  - #### **Parameters**

    - **file** (*bytes.Buffer): The body of file.

  - #### **Return Values**

    - Instance from **image.Image** type which is the file after converted to image object.
    - Instance from **error** type to indicate if there is an error or not, the value will be nil if there is no error.

- **InitSecure**

  - **Description**

    It sets insecure flag based on user input

  - #### **Parameters**

    - **VerifyServerCert** (bool): The body of file.

  - #### **Return Values**

    - Instance from **bool** indicates to is the scure flag is true or false.

- **GetMimeExtension**

  - **Description**

    It returns the mime type extension of the data

  - #### **Parameters**

    - **data** ([]byte): The file in bytes.
    - **contentType** (string): The content type of the data.
    - **filename** (string): the name of the data.

  - #### **Return Values**

    - Instance from **string** indicates to is the mime type extension of the data.

- **LogHTTPMsgHeaders**

  - **Description**

    It returns the HTTP message headers as a map.

  - #### **Parameters**

    - **methodName** (string): possible values are (**REQMOD** and **RESPMOD**).

  - #### **Return Values**

    - Instance from **map[string]interface{}** contains the header of the HTTP message.
