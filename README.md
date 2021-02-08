# go download example

## download
```
curl -L -X POST '192.168.0.166:1018/download' \
-H 'Content-Type: application/json' \
--data-raw '{"link":"http://down10.zol.com.cn/yasuo/winrarx64600.exe"}'
```

## get state
```
curl -L -X GET '192.168.0.166:1018/state'
```
