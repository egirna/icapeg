## REQMOD

**Go-ICAP-server** in **REQMOD** supports different values of **Content-Type** HTTP header:

- First value which represents a regular file sent in body like and in this case the Content-Type value maybe **application/pdf**, **text/plain** .. etc. In this type the body of the **HTTP** request which is included inside the **ICAP** request contains the file which wanted to be scanned only, **[Filebin](https://filebin.net/)** is one of the websites which use this types of **Content-Type**s in uploading files.

  ![](./img/Filebin%20contentType.png)

  ![](./img/Filebin%20requestBody.png)

- In the second type the **Content-Type** value is **multipart/form-data**, In this type the HTTP request body contains multiple parts of fields and the files, **[File.io](https://www.file.io/)** is one of the websites which use **multipart/form-data** in uploading files.

  ![](./img/fileio%20contentType.png)

  This screenshot shows different parts in the HTTP request body and the boundary -----------------------------2145730498846892413066303047 differentiates between all parts.

  ![](./img/fileio%20requsetBody.png)

- In the third type the **Content-Type** value is **application/json**, In this type the HTTP request body contains the file which wanted to be scanned encoded in base64 or a normal JSON file. **[Glasswall solutions](https://www.glasswallsolutions.com/test-drive/)** is one of the websites which encodes the file in Base64 and put it in the HTTp request body as a JSON file.

  ![](./img/gw%20contentType.png)

  ![](./img/gw%20requestBody.png)
