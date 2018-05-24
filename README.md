# ProductAIStandalone

A standalone http server for ProductAI search.

Install:

```
go get -v -u github.com/caiguanhao/ProductAIStandalone
```

Help:

```
Usage of ProductAIStandalone:
  -access-key-id string
        Access Key ID
  -listen string
        Listen To Address (default "127.0.0.1:8080")
  -service-id string
        Service ID
  -url-prefix string
        URL Prefix
```

Docker Image:

```
(echo "FROM golang:1.7.1" && \
 echo "RUN go get github.com/caiguanhao/ProductAIStandalone" && \
 echo 'ENTRYPOINT ["/go/bin/ProductAIStandalone"]') | docker build -t product-ai-standalone -
```

Docker Container:

```
docker run --name product-ai --restart always -d -p="127.0.0.1:55555:8080" product-ai-standalone \
  --access-key-id 00000000000000000000000000000000 \
     --url-prefix http://www.example.com/posts/ \
     --service-id xxxxxxxx \
         --listen 0.0.0.0:8080
```

Shell script:

```
#!/bin/bash

/usr/local/bin/ProductAIStandalone \
  --access-key-id 00000000000000000000000000000000 \
     --url-prefix http://www.example.com/posts/ \
     --service-id xxxxxxxx \
         --listen 127.0.0.1:55555 \
               >> /ProductAI.log 2>&1 &
```

Monit:

```
check process product-ai matching ProductAIStandalone
  if does not exist for 2 cycles then exec "/start-product-ai.sh"
```

cURL:

```
curl http://127.0.0.1:55555/SearchImageByURL -d 'url=https://upload.wikimedia.org/wikipedia/commons/thumb/2/28/Tischbank.jpg/640px-Tischbank.jpg'
```

nginx:

```
...

location = /SearchImageByURL {
    proxy_pass http://127.0.0.1:55555;
    proxy_set_header X-Real-IP $remote_addr;
}

...
```

LICENSE: MIT
