## 原版README
地址: [README.md](https://github.com/IngoKl/HTTPUploadExfil/blob/main/README.md)

## 新增
[+] 自定义端口号

[+] 自定义路径

[+] 新增token鉴权


## 使用
直接运行,端口号默认为58080,路径目录为当前目录,token为空

go run main.go -h
```bash
Usage of ./httpuploadexfil:
  -path string
    	Specify the storage path (default ".")
  -port string
    	Specify the listening port (default "58080")
  -token string
    	Specify the header token value
```

以参数启动
go run main.go -port=18080 -path=/Volumes/ -token=your_token
```bash
[+] Server Running...
[+] Settings: Addr ':18080'; Folder '/Volumes/'; Token 'your_token'
[+] Instructions: '/' directory quick upload, '/p' directory to manually build and upload files, '/l' directory gets the current folder contents
```
token鉴权:
<img width="995" alt="image" src="https://github.com/nu0l/HTTPUploadExfil/assets/54735907/e88dfa34-9828-45f0-8151-0012a3cf93cc">
