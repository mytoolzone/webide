AppName=code-server
IMAGE=xytschool/webide:0.0.24

isRunning=`docker ps -a | grep $AppName | wc -l`

echo $isRunning

if [ $isRunning -gt 0 ];
then
 echo  'stop '$AppName
 docker stop $AppName 
 echo  'rm $AppName'
 docker rm $AppName 
fi

docker run -it -d  --name $AppName --restart=on-failure:10   -p 127.0.0.1:8002:8080 \
  -v "$HOME/.config:/home/coder/.config" \
  -v "$HOME/webcode:/home/coder/code" \
  -u "$(id -u):$(id -g)" \
  -e "DOCKER_USER=$USER" \
  $IMAGE
