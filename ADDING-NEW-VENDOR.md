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

  - **Optional**

    > The **optional** variable is the variable which the service (eg: **xyz**) use for processing **HTTP** message, adding **optional** variables is up to the developer.

  For example if the name of the service is **xyz** and its vendor is **abc**, so the section will be like the following section:

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
  ## optional variables
  base_url = "http://abc/" #
  scan_endpoint = "xyz.com"
  api_key = "<api key>"
  timeout  = 300 
  fail_threshold = 2
  max_filesize = 0 
  ```

- ### **Vendor package** (abc package)

  - Add a new package named **abc** in the [service](./service) directory.

  - Add file named **config.go** in **abc** package and follow these instructions in this file.

    - Create a struct named **Abc** with the required attributes of any service to this vendor.

      - **Mandatory fields**:
        - **httpMsg**: It's an instance from [**HttpMsg**](utils/httpMessage.go) struct which groups two field (**HTTP request** and **HTTP response**). Developer can process the **HTTP message(request or response)** through this instance.
        - **serviceName**: It's service name value.
        - **methodName**: It's method name value.
      - **Optional fields**:
        - **generalFunc**: It's an instance from [**GeneralFunc**](service/services-utilities/general-functions/general-functions.go) struct which has a lot of function that may help the developer in processing **HTTP messages**.

      ```go
      type Abc struct {
          //mandatory
      	httpMsg                *utils.HttpMsg
      	serviceName            string
      	methodName             string
          //optional, it's up to you and to optional vaiables hav been added in service's section in config.toml file (you should map them with these struct fields)
          generalFunc            *general_functions.GeneralFunc     //optional helper field
          bypass_extensions = []
          process_extensions = ["*"] 
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

      - **doOnce**: It's a variable from [**sync**](https://pkg.go.dev/sync) package. It's used to apply singular pattern to the service's struct, to read service's **config.toml** file section only one time once **ICAPeg** runs.
      - **abcConfig**: It's an instance from **Abc** struct to store service's **config.toml** file section variables and store them in memory through it.

      ```go
      var doOnce sync.Once
      var echoConfig *Abc
      ```

    - Add **InitAbcConfig** function:

      It's used to read service's **config.toml** file section **optional** variables  and store them in memory using **abcConfig** instance. 

    - Add function named **NewAbcService** which create service from abc vendor.

      It extracts service configuration from **abcConfig** variable.

      ```go
      func NewAbcService(serviceName, methodName string, httpMsg *utils.HttpMsg) *abc {
      	return &Abc{
              	//mandatory
      		httpMsg:                httpMsg,
      		serviceName:            serviceName,
      		methodName:             methodName,
              	//optional
      		generalFunc:            general_functions.NewGeneralFunc(httpMsg),  //optional helper 
      		maxFileSize:            abcConfig.maxFileSize,
      		bypassExts:             abcConfig.bypassExts,
      		processExts:            abcConfig.processExts,
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
      func (reciever *Abc) Processing() (int, interface{}, map[string]string) {
      	// your implementation
      }
      ```

- ### [**service.go**](service/service.go)

  - Add a name of the new vendor as a constant variable in [service.go](service/servoce.go) at the start of the file in the constants section.

```go
//Vendors names
const (
	VendorEcho      = "echo"
	VendorAbc   	= "abc"
)
```

- Add a case in the switch case in **GetService** function by the new vendor

  ```go
  func GetService(vendor, serviceName, methodName string, httpMsg *utils.HttpMsg) Service {
  	switch vendor {
  	case VendorEcho:
  		return echo.NewEchoService(serviceName, methodName, httpMsg)
  	case VendorABC:
  		return abc.NewAbcService(serviceName, methodName, httpMsg)
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
