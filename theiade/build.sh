CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags  '-w -s  -extldflags "-static"' -o ./server .
tag=0.0.22
docker buildx build --platform linux/amd64 -t xytschool/fetch-code:${tag} . -f Dockerfile.fast
docker push xytschool/fetch-code:${tag}
# docker run -it --rm  xytschool/faas-api:0.0.2