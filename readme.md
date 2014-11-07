#embres
##Embed Resources into HTML file.
#### Embed resources (CSS/Javascript/Image) files into HTML page in order to create a self-contained single html file.

Saving web pages by web browser usually consists of two parts: one html file, one folder with resource files like css, javascript or images inside.  This way might not be convenient for transfering or organizing. Self-contained single file formats like ".mht" are not widely supported, so I wrote this small utility to embed resources into single html file, which is well supported by any modern web browsers.  

### Build
```sh
go get github.com/meoow/embres
```

### Usage
```sh
./embed [-i] file1.html [file2.html ...]
# use -i changes files inplace, otherwise will write to stdout.
```
