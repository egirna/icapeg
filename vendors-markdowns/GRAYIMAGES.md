# Steps to compile the ICAP when using grayimages service 


## 1. Install libwebp
#### MacOS:
```bash
brew install webp
```
#### Linux:
```bash
sudo apt-get update
sudo apt-get install libwebp-dev
```

## 2. Build ICAP
Build **ICAPeg** binary

```bash
go build .
```

Execute the file like you would for any other executable according to your OS, for Unix-based users though

```bash
./icapeg
```