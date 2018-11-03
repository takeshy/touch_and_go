# touch_and_go - File watcher

## Abstract

touch_and_go is a file watcher written in Golang. touch_and_go watches files and touch when one of these files is created, updated or removed.
I use https://github.com/yosssi/goat as reference. thanks yosssi.

## Use cases

You can use touch_and_go to Docker for Windows on shared directory.
because currently, inotify does not work on Docker for Windows. 
execute touch_and_go on docker container and if file changed, touch_and_go touch that file for iNotify.
So App as Rails, webpack-dev-server etc. can detect file changes.

## Installation

### Binary for Linux X64

https://s3-ap-northeast-1.amazonaws.com/takesy-work/touch_and_go

### From source codes

```sh
$ go get github.com/takeshy/touch_and_go
```
```sh
$ wget https://raw.githubusercontent.com/takeshy/touch_and_go/master/main.go
$ go build
```

## Configuration file

The JSON file looks like the following:

```json
{
  "watchers": [
    {
      "directory": "/home/ubuntu/my_project",
      "excludes": ["node_modules"],
    }
  ]
}
```

* `watchers` defines an array of file watchers. Each watcher definition has the following properties:
  * `excludes` (optional)
  * `directory` (optional)
* `excludes` defines an array of file names which is out of watching range.
* `directory` defines absolute directory path or the subdirectory. touch_and_go watches all files in and under this directory.  Defaults to current directory, if not specified.

## Execution

On the your project root directory execute the following command:

```sh
$ touch_and_go -c config.json -i 3000 &
```
-i An interval(ms) of a watchers default: 3000
-c config file default: config.json
