<h1 align="center">Extracted ICAP RFC</h1>



This markdown file contains the important and necessary [ICAP](https://datatracker.ietf.org/doc/html/rfc3507)'s information, This information was extracted from ICAP's [rfc](https://datatracker.ietf.org/doc/html/rfc3507).

Sections From [1](https://datatracker.ietf.org/doc/html/rfc3507#section-1) until [3](https://datatracker.ietf.org/doc/html/rfc3507#section-3) contains introduction, terminologies and how it works, you can find read it from the [rfc](https://datatracker.ietf.org/doc/html/rfc3507):

- [section 1](https://datatracker.ietf.org/doc/html/rfc3507#section-1).
- [section 2](https://datatracker.ietf.org/doc/html/rfc3507#section-2).
- [section 3](https://datatracker.ietf.org/doc/html/rfc3507#section-3).

# Table of contents

- [4. Protocol semantics](#protocolSemantics)

  - [4.1 General Operation](#generalOperation)
  - [4.2 ICAP URIs](#icapUri)
  - [4.3 ICAP Headers](#icapHeaders)
    - [4.3.1 Headers Common to Requests and Responses](#headersCommonToRequestsAndResponses)
    - [4.3.2 Request Headers](#requestHeaders)
    - [4.3.3 Response Headers](#responseHeaders)
  - [4.5 Message Preview](#messagePreview)
  - [4.6 Responses outside of Previews](#responsesOutsideOfPreviews)
  - [4.7 ISTag Response Header](#ISTagResponseHeader)
  - [4.7 Request Modification Mode](#requestModMode)
  - [4.8 Request Modification Mode](#requestModMode)
    - [4.8.1 Request](#requestReq)
    - [4.8.2 Response](#responseReq)
    - [4.8.3 Examples](#examplesReq)
  - [4.9 Response Modification Mode](#respModMode)
    - [4.9.1 Request](#requestResp)
    - [4.9.2 Response](#responseResp)
    - [4.9.3 Examples](#examplesResp)
  - [4.10 OPTIONS Method](#optionsMethod)
    - [4.10.1 OPTIONS Request](#optionsRequest)
    - [4.10.2 OPTIONS Response](#optionsResponse)
    - [4.10.3 OPTIONS Examples](#optionsExamples)

- [5. Caching](#caching)

- [6. Implementation Notes](#implementationNotes)

  - [6.1 Vectoring points](#vectoringPoints)
  - [6.2 Application Level Errors](#applicationLevelErrors)
  - [6.3  Use of Chunked Transfer-Encoding](#UseofChunkedTransfer-Encoding)
  - [6.4 Distinct URIs for Distinct Services](#DistinctURIsforDistinctServices)

- [7.  Security Considerations](#securityConsiderations)

  - [7.1 Authentication](#auth)
  - [7.2 Encryption](#encryption)
  - [7.3 Service Validation](#serviceValidation)

- [8.0 Motivations and Design Alternatives](#MotivationsandDesignAlternatives)

  - [8.2 Mandatory Use of Chunking](#MandatoryUseofChunking)
  - [8.3 Use of the null-body directive in the Encapsulated header](#Useofthenull-bodydirectiveintheEncapsulatedheader)

  

  

### 4. Protocol semantics <a name="protocolSemantics"></a>

- ### 4.1 General Operation <a name="generalOperation"></a>

  - **ICAP** uses **TCP/IP** as a transport protocol.
  - The default port is **1344**, but other ports may be used.
  - Requests and responses use the generic message format of [RFC 2822](https://datatracker.ietf.org/doc/html/rfc2822) [3](https://datatracker.ietf.org/doc/html/rfc3507#ref-3) -- that is, a start-line (either a   request line or a status line), a number of header fields (also known   as "headers"), an empty line (i.e., a line with nothing preceding the **CRLF**) indicating the end of the header fields, and a message-body.
  - The header lines of an **ICAP** message specify the **ICAP** resource being
       requested.
  - A single transport connection MAY (perhaps even **SHOULD**) be re-used for multiple request/response pairs.
  - Requests are matched up with responses by allowing only one outstanding request on a transport connection at a time.
  - Multiple parallel connections MAY be used as in **HTTP**.

  

- ### 4.2 ICAP URI <a name="icapUri"></a>

  - **ICAP URI** example:

    ```
    icap://icap.example.net:2000/services/icap-service-1
    ```

  - If the port is empty or not given, port **1344** is assumed.

  - Any arguments that an **ICAP** client wishes to pass to an **ICAP** service to modify the nature of the service **MAY** be passed as part of the **ICAP-URI**, using the standard "**?**", **example**: "icap://icap.net/service?mode=translate&lang=french".

  

- ### 4.3 ICAP Headers <a name="icapHeaders"></a>

  - User-defined header extensions are allowed. all user-defined headers **MUST** follow the **"X-"** naming convention **("X-Extension-Header: Foo")**,**ICAP** implementations MAY ignore any **"X-"** headers without loss of compliance with the protocol as defined in this document.

  -  Each header field consists of a name followed by a colon ("**:**") and the field value, Field names are case-insensitive.

  - #### 4.3.1 Headers Common to Requests and Responses <a name="headersCommonToRequestsAndResponses"></a>

    - The headers of all **ICAP** messages MAY include the following directives: (Cache-Control, Connection, Date, Expires, Pragma, Trailer, Upgrade).
    -  "Transfer-Encoding" option is not allowed.
    -  The Upgrade header MAY be used to negotiate Transport-Layer Security on an **ICAP** connection.

  - #### 4.3.2 Request Headers <a name="requestHeaders"></a>

    - **ICAP** requests **MUST** start with a request line that contains a method, the complete **URI** of the **ICAP** resource being requested, and an **ICAP** version string, The current version number of **ICAP** is "**1.0**".
    -  This version of ICAP defines three methods: (**REQMOD**  - for Request Modification, **RESPMOD** - for Response Modification, **OPTIONS** - to learn about configuration). 
    - The **OPTIONS** method **MUST** be implemented by all **ICAP** servers, All other methods are optional and **MAY** be implemented.
    - User-defined extension methods are allowed. Before attempting to use   an extension method, an **ICAP** client **SHOULD** use the **OPTIONS** method to query the **ICAP** server's list of supported methods.
    - If an **ICAP** server receives a request for an unknown method, it **MUST** give a **501** error response.
    - A well-formed **ICAP** request line looks like the following example:                                                                  **RESPMOD** icap://icap.example.net/translate?mode=french ICAP/1.0.
    - Request specific headers are alllowed in **ICAP**: Authorization, Allow, From, Host (REQUIRED in ICAP as it is in HTTP/1.1), Referer, User-Agent.
    - Request headers unique to **ICAP** defined: Preview

  - #### 4.3.3 Response Headers<a name="responseHeaders"></a>

    - **ICAP** responses MUST start with an **ICAP** status line, similar in form   to that used by **HTTP**, including the **ICAP** version and a status code.                                                                                                                    For example: **ICAP/1.0 200 OK**

    - Semantics of **ICAP** status codes in **ICAP** match the status codes definedby **HTTP**.                                  **ICAP** error codes that differ from their HTTP counterparts are: 

      **100** - Continue after ICAP Preview.

      **204** - No modifications needed. 

      **400** - Bad request.

      **404** - ICAP Service not found.

      **405** - Method not allowed for service (e.g., **RESPMOD** requested for service that supports only **REQMOD**).

      **408** - Request timeout. **ICAP** server gave up waiting for a request from an ICAP client.     

      **500** - Server error.  Error on the **ICAP** server, such as "out of disk space".

      **501** - Method not implemented. This response is illegal for an **OPTIONS** request since implementation of **OPTIONS** is mandatory.

      **502** - Bad Gateway. This is an ICAP proxy and proxying produced an error.              

      **503** - Service overloaded.  The ICAP server has exceeded a maximum  connection limit associated with this service; the ICAP client should not exceed this limit in the future. 

    -  **ICAP**'s response-header fields allow the server to pass additional information in the response that cannot be placed in the **ICAP**'s status line.

    - A response-specific header is allowed in **ICAP** requests, following the same semantics as the corresponding **HTTP** response headers. This is: Server.

    - In addition to **HTTP-like** headers, there is also a response header unique to **ICAP** defined: **ISTag**.

    

- ### 4..5 Message Preview<a name="messagePreview"></a>

  - **ICAP** servers **SHOULD** use the **OPTIONS** method to specify how many bytes of preview are needed for a particular **ICAP** application on a per-resource basis. Clients **SHOULD** be able to provide Previews of at least **4096** bytes. Clients furthermore **SHOULD** provide a Preview when using any **ICAP** resource that has indicated a Preview is useful. (This indication might be provided via the **OPTIONS** method, or some other "out-of-band" configuration.) Clients **SHOULD NOT** provide a larger Preview than a server has indicated it is willing to accept.

  -  To effect a Preview, an **ICAP** client **MUST** add a "**Preview:**" header to its request headers indicating the length of the preview. client then sends:

    - all of the encapsulated header sections.
    - the beginning of the encapsulated body section, if any, up to the number of bytes advertised in the Preview (**possibly 0**).

  -  After the Preview is sent, the client stops and waits for ani ntermediate response from the **ICAP** server before continuing. This mechanism is similar to the "**100-Continue**" feature found in **HTTP**, except that the stop-and-wait point can be within the message body. In contrast, **HTTP** requires that the point must be the boundary   between the headers and body.

  - After sending the preview, the **ICAP** client will wait for a response from the **ICAP** server. The response **MUST** be one of the following:

    - **204 No Content.** The ICAP server does not want to (or can not) modify the ICAP client's request. The ICAP client MUST treat this the same as if it had sent the entire message to the ICAP server and an identical message was returned.
    - **ICAP** reqmod or respmod response, depending what method was the original request.
    - **100 Continue.** If the entire encapsulated **HTTP** body did not fit in the preview, the **ICAP** client **MUST** send the remainder of its **ICAP** message, starting from the first chunk after the preview.  If the entire message fit in the preview (detected by the "**EOF**"  symbol explained below), then the **ICAP** server **MUST NOT **respond with **100 Continue**. 

  - For example, to effect a Preview consisting of only encapsulated **HTTP** headers, the **ICAP** client would add the following header to the **ICAP** request:       

    ```
    Preview: 0
    ```

    This indicates that the **ICAP** client will send only the encapsulated header sections to the **ICAP** server, then it will send a zero-length chunk and stop and wait for a "go ahead" to send more encapsulated body bytes to the **ICAP** server. 

  - We define an **HTTP** chunk-extension of "**ieof**" to indicate that an **ICAP** chunk is the last chunk. The **ICAP** server **MUST** strip this chunk extension before passing the chunk data to an **ICAP** application process.

  - In another example, if the preview is 1024 bytes and the origin response is **1024** bytes in two chunks, then the encapsulation would appear as follows:       

    ```
    200\r\n
    
    <512 bytes of data>\r\n
    
    200\r\n
    
    <512 bytes of data>\r\n
    
    0; ieof\r\n\r\n
    
    <204 or modified response> (100 Continue disallowed due to ieof) 
    ```

  -  If the preview is 1024 bytes and the origin response is **1025** bytes (and the **ICAP** server responds with **100-continue**), then these chunks would appear on the wire:

    ```
    200\r\n      
    
    <512 bytes of data>\r\n
    
    200\r\n      
    
    <512 bytes of data>\r\n
    
    0\r\n       
    
    <100 Continue Message>       
    
    1\r\n      
    
    <1 byte of data>\r\n      
    
    0\r\n\r\n  //no ieof because we are no longer in preview mode 
    ```

    

- ### **4.6 "204 No Content" Responses outside of Previews**  <a name="responsesOutsideOfPreviews"></a>

  - An **ICAP** client **MAY** include "**Allow: 204**" in its request headers, indicating that the server **MAY** reply to the message with a "**204 No Content**" response if the object does not need modification. An **ICAP** client **MAY** choose to honor "**204 No Content**" responses for an   entire message. This is the decision of the client because it imposes a burden on the client of buffering the entire message. 

  - If an **ICAP** server receives a request that does not have "**Allow: 204**", it **MUST NOT** reply with a **204**. In this case, an **ICAP** server **MUST** return the entire message back to the client, even though it is identical to the message it received.

  -  In case of message preview an **ICAP** server can respond with a **204 No Content** message in response to a   message preview **EVEN** if the original request did not have the "**Allow: 204**" header.

    

- ### 4.7 ISTag Response Header  <a name="ISTagResponseHeader"></a>

- It is a 32-byte-maximum alphanumeric string of data (not including the null character).

- An **ISTag** validates that previous **ICAP** server responses can still be considered fresh by an **ICAP** client that   may be caching them. If a change on the **ICAP** server invalidates previous responses, the **ICAP** server can invalidate portions of the   ICAP client's cache by changing its **ISTag**. The **ISTag MUST** be included in every **ICAP** response from an **ICAP** server.

- For example, consider a virus-scanning **ICAP** service. The **ISTag** might be a combination of the virus scanner's software version and the release number of its virus signature database.  When the database is updated, the **ISTag** can be changed to invalidate all previous responses that had been certified as "**clean**" and cached with the old **ISTag**.

-  an **ISTag** validates all entities generated by a particular service (**URI**).  A change in the **ISTag** invalidates all the other entities provided a service with the old **ISTag**, not just the entity whose response contained the updated **ISTag**.

-  The syntax of an **ISTag** is simply:     

  ```
   ISTag = "ISTag: " quoted-string
  ```

  ​    For example:      

  ```
  ISTag: "874900-1994-1c02798"
  ```

  

- ### 4.8 Request Modification Mode  <a name="requestModMode"></a>

  - #### 4.8.1 Request <a name="requestReq"></a>

    In **REQMOD** mode, the **ICAP** request **MUST** contain an encapsulated **HTTP** request. The headers and body (**if any**) **MUST** both be encapsulated, except that hop-by-hop headers are not encapsulated. 

  - #### 4.8.2 Response <a name="responseReq"></a>

    The response from the **ICAP** server back to the **ICAP** client may take   one of four forms:    

    - An error indication,

    - A **204** indicating that the **ICAP** client's request requires no adaptation.
    -  An encapsulated, adapted version of the **ICAP** client's request, or
    - An encapsulated **HTTP** error response. Note that Request Modification requests may only be satisfied with **HTTP** responses incases when the **HTTP** response is an error (e.g., **403 Forbidden**). 

  - #### 4.8.3 Examples <a name="examplesReq"></a>

    You can find examples [here](https://datatracker.ietf.org/doc/html/rfc3507#section-4.8.3).

    

- ### 4.9 Response Modification Mode <a name="respModMode"></a>

  An ICAP client sends an   origin server's **HTTP** response to an ICAP server, and (if available)   the original client request that caused that response.

  - #### 4.9.1 Request <a name="requestResp"></a>

    the header and body of the **HTTP** response to be modified **MUST** be included in the ICAP body.   If available, the header of the original client request **SHOULD** also be included. As with the other method, the hop-by-hop headers of the encapsulated messages **MUST NOT** be forwarded. The Encapsulated header **MUST** indicate the byte-offsets of the beginning of each of these fou parts.

  - #### 4.9.2 Response <a name="responseResp"></a>

    - The response from the **ICAP** server looks just like a reply in the Request Modification method, that is
      - An error indication.
      - An encapsulated and potentially modified **HTTP** response header and response body, or
      - An **HTTP** response 204 indicating that the **ICAP** client's request requires no adaptation.
    - If the return code is a **2XX**, the **ICAP** client **SHOULD** continue its normal execution of the response. The   **ICAP** client **MAY** re-examine the headers in the response's message headers in order to make further decisions about the response (e.g.,   its cachability).

  - #### 4.9.3 Examples <a name="examplesResp"></a>

    You can find examples [here](https://datatracker.ietf.org/doc/html/rfc3507#section-4.9.3).

    

- ### 4.10 OPTIONS Method <a name="optionsMethod"></a>

  The **ICAP** "**OPTIONS**" method is used by the **ICAP** client to retrieve configuration information from the **ICAP** server.  In this method, the **ICAP** client sends a request addressed to a specific **ICAP** resource and receives back a response with options that are specific to the service named by the **URI**.  All **OPTIONS** requests **MAY** also return   options that are global to the server (i.e., apply to all services). 

  - #### 4.10.1 OPTIONS Request <a name="optionsRequest"></a>

    The **OPTIONS** method consists of a request-line  such as the following example:

    ```
    OPTIONS icap://icap.server.net/sample-service ICAP/1.0 User-Agent: ICAP-client-XYZ/1.001
    ```

  - #### 4.10.2 OPTIONS Response <a name="optionsResponse"></a>

  - The **OPTIONS** response consists of a status line followed by a series of header field **names-value** pairs   optionally followed by an **opt-body**. Multiple values in the value field **MUST** be separated by commas. If an opt-body is present in the **OPTIONS** response, the **Opt-body-type** header describes the format of the **opt-body**. 

    The **OPTIONS** headers supported in this version of the protocol are:

    - Methods:

      The method that is supported by this service.  This header MUST be included in the OPTIONS response.  The **OPTIONS** method **MUST NOT** be in the Methods' list since it **MUST** be supported by all the **ICAP**      server implementations. Each service should have a distinct URI and support only one method in addition to **OPTIONS** For example:      

      ```
      Methods: RESPMOD
      ```

    - Service:

      A text description of the vendor and product name. This header **MAY** be included in the **OPTIONS** response. For example:      

      ```
      Service: XYZ Technology Server 1.0
      ```

    - ISTag:

      For example:      

      ```
      ISTag: "5BDEEEA9-12E4-2" 
      ```

    - Encapsulated:

      For example:      

      ```
      Encapsulated: opt-body=0
      ```

    - Opt-body-type:

      A token identifying the format of the opt-body.  (Valid opt-body types are not defined by ICAP.) This header MUST be included in the **OPTIONS** response **ONLY** if an opt-body type is present. For example:      

      ```
      Opt-body-type: XML-Policy-Table-1.0
      ```

    - Max-Connections:

      The maximum number of **ICAP** connections the server is able to support. This header **MAY** be included in the **OPTIONS** response. For example:      

      ```
      Max-Connections: 1500
      ```

    - Options-TTL:

      The time (in seconds) for which this **OPTIONS** response is valid. If none is specified, the **OPTIONS** response does not expire. This header **MAY** be included in the OPTIONS response. The **ICAP** client      MAY reissue an **OPTIONS** request once the **Options-TTL** expires. For example:

      ```      
      Options-TTL: 3600
      ```

    - Date: 

      The server's clock, specified as an RFC 1123 compliant date/time string. This header MAY be included in the **OPTIONS** response. For example:      

      ```
      Date: Fri, 15 Jun 2001 04:33:55 GMT
      ```

    - Service-ID: 

      A short label identifying the ICAP service. It **MAY** be used in attribute header names. This header MAY be included in the **OPTIONS** response. For example:      

      ```
      Service-ID: xyztech
      ```

    - Allow:

      A directive declaring a list of optional ICAP features that this server has implemented. This header **MAY** be included in the **OPTIONS** response. In this document we define the value "204" to indicate that the ICAP server supports a 204 response. For example:      

      ```
      Allow: 204
      ```

    - Preview:

      The number of bytes to be sent by the ICAP client during a preview. This header **MAY** be included in the **OPTIONS** response. For example:      

      ```
      Preview: 1024
      ```

    - Transfer-Preview:

      A list of file extensions that should be previewed to the **ICAP** server before sending them in their entirety.  This header **MAY** be included in the **OPTIONS** response. Multiple file extensions values should be separated by commas. The wildcard value "*****" specifies the default behavior for all the file extensions not specified inany other **Transfer-*** header (see below). For example:      

      ```
      Transfer-Preview: *
      ```

    - Transfer-Ignore:

      A list of file extensions that should **NOT** be sent to the **ICAP** server. This header **MAY** be included in the **OPTIONS** response. Multiple file extensions should be separated by commas. For example:      

      ```
      Transfer-Ignore: html
      ```

    - Transfer-Complete:

      A list of file extensions that should be sent in their entirety (without preview) to the ICAP server. This header **MAY** be included in the **OPTIONS** response.  Multiple file extensions values should be separated by commas. For example: 

      ```
      Transfer-Complete: asp, bat, exe, com, ole    
      ```

      Note: If any of **Transfer-*** are sent, exactly one of them MUST contain   the wildcard value "*****" to specify the default.  If no **Transfer-*** are sent, all responses will be sent in their entirety (without Preview).

  - #### 4.10.3 OPTIONS Examples <a name="optionsExamples"></a>

    You can find examples [here](https://datatracker.ietf.org/doc/html/rfc3507#section-4.10.3).

    

### 5. Caching <a name="caching"></a>

- **ICAP** servers' responses **MAY** be cached by **ICAP** clients, just as any other surrogate might cache **HTTP** responses.
- In Request Modification mode, the **ICAP** server **MAY** include caching directives in the **ICAP** header section of the **ICAP** response (**NOT** in the encapsulated **HTTP** request of the **ICAP** message body). 
- In Response Modification mode, the **ICAP** server **MAY** add or modify the **HTTP** caching directives located in the encapsulated **HTTP** response (**NOT** in the **ICAP** header section). 
- Consequently, the **ICAP** client **SHOULD** look for caching directives in the **ICAP** headers in case of **REQMOD**, and in the encapsulated **HTTP** response in case of **RESPMOD**.
- In cases where an **ICAP** server returns a modified version of an object created by an origin server, such as in Response Modification mode,   the expiration of the **ICAP-modified** object **MUST NOT** be longer than   that of the origin object. In other words, ICAP servers **MUST NOT** extend the lifetime of origin server objects, but **MAY** shorten it.
- In cases where the **ICAP** server is the authoritative source of an **ICAP** response, such as in Request Modification mode, the **ICAP** server is not restricted in its expiration policy. 
- Note that the **ISTag** response-header may also be used to providing caching hints to clients.



### 6. Implementations Notes  <a name="implementationNotes"></a>

- #### 6.1 Vectoring points <a name="vectoringPoints"></a>

  1. Adaptation of client requests. This is adaptation done every time a request arrives from a client. This is adaptation done when a request is "on its way into the cache". Factors such as the state of the objects currently cached will determine whether or not this request actually gets forwarded to an origin server (instead of, say, getting served off the cache's disk). 
  2.  Adaptation of requests on their way to an origin server. Although this type of adaptation is also an adaptation of       requests similar to (**i**), it describes requests that are "**on their way out of the cache**"; i.e., if a request actually requires tha an origin server be contacted. These adaptation requests are not necessarily specific to particular clients. An example would be addition of "**Accept:**"  headers for special devices; these adaptations can potentially apply to many clients.
  3. Adaptations of responses coming from an origin server. This is the adaptation of an object "**on its way into the cache**". In other words, this is adaptation that a surrogate might want to perform on an object before caching it.  The adapted object may subsequently served to many clients.  An example of this type of adaptation is virus checking: a surrogate will want to check an incoming origin reply for viruses once, before allowing it into the cache -- not every time the cached object is served to a client.

- #### 6.2 Application Level Errors <a name="applicationLevelErrors"></a>

  ```
  Error name                                     Value
     ====================================================
     ICAP_CANT_CONNECT                               1000
     ICAP_SERVER_RESPONSE_CLOSE                      1001
     ICAP_SERVER_RESPONSE_RESET                      1002
     ICAP_SERVER_UNKNOWN_CODE                        1003
     ICAP_SERVER_UNEXPECTED_CLOSE_204                1004
     ICAP_SERVER_UNEXPECTED_CLOSE                    1005
  
     1000 ICAP_CANT_CONNECT:
         "Cannot connect to ICAP server".
  
         The ICAP server is not connected on the socket.  Maybe the ICAP
         server is dead or it is not connected on the socket.
  
     1001 ICAP_SERVER_RESPONSE_CLOSE:
         "ICAP Server closed connection while reading response".
  
         The ICAP server TCP-shutdowns the connection before the ICAP
         client can send all the body data.
  
     1002 ICAP_SERVER_RESPONSE_RESET:
         "ICAP Server reset connection while reading response".
  
         The ICAP server TCP-reset the connection before the ICAP client
         can send all the body data.
  
     1003 ICAP_SERVER_UNKNOWN_CODE:
         "ICAP Server sent unknown response code".
  
         An unknown ICAP response code (see Section 4.x) was received by
         the ICAP client.
  
     1004 ICAP_SERVER_UNEXPECTED_CLOSE_204:
         "ICAP Server closed connection on 204 without 'Connection: close'
         header".
  
         An ICAP server MUST send the "Connection: close" header if
         intends to close after the current transaction.
  
     1005 ICAP_SERVER_UNEXPECTED_CLOSE:
         "ICAP Server closed connection as ICAP client wrote body
         preview".
  ```

- #### 6.3  Use of Chunked Transfer-Encoding <a name="UseofChunkedTransfer-Encoding"></a>

  **ICAP** messages **MUST** use the "**chunked**" **transfer-encoding** within the encapsulated body section as defined in **HTTP/1.1**. This requires that **ICAP** client implementations convert incoming objects "**on the fly**" to chunked from whatever transfer-encoding on which they arrive. However, the transformation is simple:

  - For objects arriving using "**Content-Length**" headers, one big chunk can be created of the same size as indicated in the **Content-Length** header.
  - For objects arriving using a **TCP** close to signal the end of the object, each incoming group of bytes read from the **OS** can be converted into a chunk (by writing the length of the bytes read, followed by the bytes themselves)   
  - For objects arriving using chunked encoding, they can be retransmitted as is (without re-chunking).

- #### 6.4 Distinct URIs for Distinct Services <a name="DistinctURIsforDistinctServices"></a>

  **ICAP** servers **SHOULD** assign unique **URIs** to each service they provide, even if such services might theoretically be differentiated based on their method. In other words, a **REQMOD** and **RESPMOD** service should never have the same **URI**, even if they do something that is conceptually the same.



### 7. Security Considerations <a name="securityConsiderations"></a>

- #### 7.1 Authentication <a name="auth"></a>

  Authentication in ICAP is very similar to proxy authentication in HTTP. Specifically, the following rules apply:

  - WWW-Authenticate challenges and responses are for end-to-end authentication between a client (user) and an origin server. As any proxy, ICAP clients and ICAP servers MUST forward these headers without modification.
  -  If authentication is required between an ICAP client and ICAP server, hop-by-hop Proxy Authentication MUST be used. 

- #### 7.2 Encryption 

  Users of ICAP should note well that ICAP messages are not encrypted for transit by default.  In the absence of some other form o encryption at the link or network layers, eavesdroppers may be able to record the unencrypted transactions between ICAP clients and servers. The Upgrade header MAY be used to negotiate transport-layer security for an ICAP connection. Note also that end-to-end encryption between a client and origin server is likely to preclude the use of value-added services by intermediaries such as surrogates. An ICAP server that is unable to   decrypt a client's messages will, of course, be unable to perform any transformations on it. 

- #### 7.3 Service Validation <a name="serviceValidation"></a>

  Normal HTTP surrogates, when operating correctly, should not affect the end-to-end semantics of messages that pass through them. This forms a well-defined criterion to validate that a surrogate is working correctly: a message should look the same before the surrogate as it does after the surrogate. In contrast, ICAP is meant to cause changes in the semantics of messages on their way from origin servers to users. The criteria for a correctly operating surrogate are no longer as easy to define.This will make validation of ICAP services significantly more   difficult. Incorrect adaptations may lead to security vulnerabilities that were not present in the unadapted content.



### 8. Motivations and Design Alternatives <a name="MotivationsandDesignAlternatives"></a>

- #### 8.2 Mandatory Use of Chunking <a name="MandatoryUseofChunking"></a>

  - Chunking is mandatory in ICAP encapsulated bodies he chunked encoding allows boththe client and server to keep the transport-layer connection open for later reuse.
  - ICAP servers (and their developers) should be encouraged to produce "incremental" responses where possible, By standardizing on a single encapsulation mechanism, we avoid the complexity that would be required in client and server software to support multiple mechanisms.
  - While chunking of encapsulated bodies is mandatory, encapsulated headers are not chunked.

- #### 8.3 Use of the null-body directive in the Encapsulated header<a name="Useofthenull-bodydirectiveintheEncapsulatedheader"></a>

  parsers do not know in advance how much header data is coming (e.g., for buffer allocation). ICAP does not allow chunking in the header part. To compensate, the "null-body" directive allows the final header's length to be   determined, despite it not being chunked.











