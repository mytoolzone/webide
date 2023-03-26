#CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags  '-w -s  -extldflags "-static"' -o ./server .
tag=0.0.22
APP=webide
docker buildx build --platform linux/amd64 -t xytschool/${APP}:${tag} . -f Dockerfile
docker push xytschool/${APP}:${tag}
# docker run -it --rm  xytschool/${webide}:0.0.2