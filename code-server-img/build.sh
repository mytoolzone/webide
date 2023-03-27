#CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags  '-w -s  -extldflags "-static"' -o ./server .
tag=0.0.24
APP=webide
docker build -t xytschool/${APP}:${tag} . -f Dockerfile
docker push xytschool/${APP}:${tag}
# docker run -it --rm  xytschool/${webide}:0.0.2
