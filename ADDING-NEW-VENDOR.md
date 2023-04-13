## **How to add a new service for a new vendor?**

Assume that the name of the vendor is **abc** and the name of the service **xyz**, Let's start with the steps: 

- ### [**Config.toml**](config.toml)

  > **Note**: You will find an explaination for **config.toml** section in [**README.md**](README.md).

  You should add a section with the name of the service in [**config.toml**](./config.toml).

  You should add the required variables of the service. There are **two** types of variables in **config.toml** file:

  - **Mandatory**:

    > The **mandatory** variable is the variable which **ICAPeg server** uses for processing **ICAP** message.

    - **Mandatory variables**:
      - **vendor**
      - **service_caption**
      - **service_tag**
      - **req_mode**
      - **resp_mode**
      - **shadow_service**
      - **preview_bytes**
      - **preview_enabled**
      - **process_extensions**
      - **reject_extensions**
      - **bypass_extensions**

  - **Optional**

    > The **optional** variable is the variable which the service (eg: **xyz**) use for processing **HTTP** message, adding **optional** variables is up to the developer.

  For example, if the name of the service is **xyz** and its vendor is **abc**, the section will be like the following section:

  ```toml
  [xyz]
  ## mandatory variables
  vendor = "abc"
  service_caption= "XYZ service"   
  service_tag = "ABC ICAP"  
  req_mode=true 
  resp_mode=true 
  shadow_service=false
  preview_bytes = "1024" 
  preview_enabled = true
  process_extensions = ["pdf", "zip", "com"] 
  reject_extensions = ["docx"]
  bypass_extensions = ["*"]
  ## optional variables
  base_url = "http://abc/" #
  scan_endpoint = "xyz.com"
  api_key = "<api key>"
  timeout  = 300 
  fail_threshold = 2
  max_filesize = 0 
  ```

- ### **Vendor package** (abc package)

  - Add a new package named **abc** in the [service](./service/services) directory.

  - Add file named **config.go** in **abc** package and follow these instructions in this file.

    - Create a struct named **Abc** with the required attributes of any service to this vendor.

      - **Mandatory fields**:
        - **xICAPMetadata**: It's the id of every **ICAP** request sent to **ICAPeg**.
        - **httpMsg**: It's an instance from [**HttpMsg**](http-message/httpMessage.go) struct which groups two field (**HTTP request** and **HTTP response**). The developer can process the **HTTP message(request or response)** through this instance.
        - **serviceName**: It's the service name value.
        - **methodName**: It's method name value.

      - **Optional fields**:
        - **generalFunc**: It's an instance from [**GeneralFunc**](service/services-utilities/general-functions/general-functions.go) struct which has a lot of function that may help the developer in processing **HTTP messages**.

      ```go
      type Abc struct {
          //mandatory
          xICAPMetadata              string
      	httpMsg                *utils.HttpMsg
      	serviceName            string
      	methodName             string
          bypassExts                 []string
      	processExts                []string
      	rejectExts                 []string
      	extArrs                    []services_utilities.Extension
      
          //optional, it's up to you and to optional variables have been added in the service section in config.toml file (you should map them with these struct fields)
          generalFunc            *general_functions.GeneralFunc     //optional helper field
          base_url = "echo" 
          scan_endpoint = "echo"
          api_key = "<api key>"
          timeout  = 300 #seconds , ICAP will return 408 - Request timeout
          fail_threshold = 2
          max_filesize = 0 #bytes
          return_original_if_max_file_size_exceeded=false
      }
      ```

    - Add **doOnce** and  **abcConfig** variables:

      - **doOnce**: It's a variable from [**sync**](https://pkg.go.dev/sync) package. It's used to apply the singular pattern to the service's struct, read the service's **config.toml** file section only one time once **ICAPeg** runs.
      - **abcConfig**: It's an instance from **Abc** struct to store the service's **config.toml** file section variables and store them in memory through it.

      ```go
      var doOnce sync.Once
      var abcConfig *Abc
      ```

    - Add **InitAbcConfig** function:

      It's used to read service's **config.toml** file section **optional** variables  and store them in memory using **abcConfig** instance. 

      ```go
      func InitEchoConfig(serviceName string) {
      	doOnce.Do(func() {
      		echoConfig = &Echo{
      			maxFileSize:                readValues.ReadValuesInt(serviceName + ".max_filesize"),
      			bypassExts:                 readValues.ReadValuesSlice(serviceName + ".bypass_extensions"),
      			processExts:                readValues.ReadValuesSlice(serviceName + ".process_extensions"),
      			rejectExts:                 readValues.ReadValuesSlice(serviceName + ".reject_extensions"),
      			BaseURL:                    readValues.ReadValuesString(serviceName + ".base_url"),
      			Timeout:                    readValues.ReadValuesDuration(serviceName+".timeout") * time.Second,
      			APIKey:                     readValues.ReadValuesString(serviceName + ".api_key"),
      			ScanEndpoint:               readValues.ReadValuesString(serviceName + ".scan_endpoint"),
      			FailThreshold:              readValues.ReadValuesInt(serviceName + ".fail_threshold"),
      			returnOrigIfMaxSizeExc:     readValues.ReadValuesBool(serviceName + ".return_original_if_max_file_size_exceeded"),
      			return400IfFileExtRejected: readValues.ReadValuesBool(serviceName + ".return_400_if_file_ext_rejected"),
      		}
      		echoConfig.extArrs = services_utilities.InitExtsArr(echoConfig.processExts, echoConfig.rejectExts, echoConfig.bypassExts)
      	})
      }
      
      ```

      

    - Add a function named **NewAbcService** which creates a service from abc vendor.

      It extracts service configuration from **abcConfig** variable.

      ```go
      func NewAbcService(serviceName, methodName string, httpMsg *utils.HttpMsg, xICAPMetadata string) *abc {
      	return &Abc{
              	//mandatory
              xICAPMetadata:              xICAPMetadata, //the id of the ICAP request
      		httpMsg:                httpMsg,
      		serviceName:            serviceName,
      		methodName:             methodName,
              bypassExts:                 echoConfig.bypassExts,
      		processExts:                echoConfig.processExts,
      		rejectExts:                 echoConfig.rejectExts,
              	//optional
      		generalFunc:            general_functions.NewGeneralFunc(httpMsg),  //optional helper 
      		maxFileSize:            abcConfig.maxFileSize,
      		extArrs:                    echoConfig.extArrs,
      		BaseURL:                abcConfig.BaseURL,
      		Timeout:                abcConfig.Timeout * time.Second,
      		APIKey:                 abcConfig.APIKey,
      		ScanEndpoint:           abcConfig.ScanEndpoint,
      		FailThreshold:          abcConfig.FailThreshold,
      		returnOrigIfMaxSizeExc: abcConfig.returnOrigIfMaxSizeExc,
      	}
      }
      ```

  - Create **abc.go** in **abc** package and follow these instructions in this file.

    - Add function **Processing** to **abc** struct to implement the interface

      ```go
      //Processing is a func used for to processing the http message
      func (reciever *Abc) Processing(partial bool) (int, interface{}, map[string]string, map[string]interface{}, map[string]interface{}, map[string]interface{}) {
      	// your implementation
      }
      ```
    
      The values **Processing() function** paramaters:
    
      - **partial bool**: Boolean variable refer to wether the content is partial content or complete content.
      
      The values **Processing() function** returns:
      
      - **int**: ICAP response status code.
      - **interface{}**: HTTP message after processing, It can be HTTP request if the ICAP request methos is REQMOD or HTTP response if the ICAP request methos is RESPMOD.
      - **map[string]string**: map contains any headers related to the vendor and wanted to be added to ICAP response.
      -  **map[string]interface{}**: map contains HTTP message headers before the service of the vendor start processing the message.
      - **map[string]interface{}**: map contains HTTP message headers after the service of the vendor started processing the message.
      - **map[string]interface{}**: map contains any information the service of the vendor wants to log.

- ### [**service.go**](service/service.go)

  - Add the name of the new vendor as a constant variable in [service.go](service/service.go) at the start of the file in the constants section.

```go
//Vendors names
const (
	VendorEcho      = "echo"
	VendorAbc   	= "abc"
)
```

- Add a case in the switch case in **GetService** function by the new vendor

  ```go
  func GetService(vendor, serviceName, methodName string, httpMsg *utils.HttpMsg, xICAPMetadata string) Service {
  	switch vendor {
  	case VendorEcho:
  		return echo.NewEchoService(serviceName, methodName, httpMsg, xICAPMetadata)
  	case VendorABC:
  		return abc.NewAbcService(serviceName, methodName, httpMsg, xICAPMetadata)
  	}
  	return nil
  }
  ```

- Add a case in the switch case in **InitServiceConfig** function by the new vendor

  ```go
  func InitServiceConfig(vendor, serviceName string) {
  	switch vendor {
  	case VendorEcho:
  		echo.InitEchoConfig(serviceName)
     	case VendorAbc:
  		echo.InitAbcConfig(serviceName)
  	}
  }
  ```

Please, check [**echo vendor**](service/services/echo/) to relate to above explanation.

Now you can run **ICAPeg** and try it with **your service**.
