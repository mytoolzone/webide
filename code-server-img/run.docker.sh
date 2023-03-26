isRunning=`docker ps -a | grep code-server | wc -l`
echo $isRunning

if [ $isRunning -gt 0 ];
then
 echo  'stop code-server'
 docker stop code-server
 echo  'rm code-server'
 docker rm code-server
fi


docker run -it -d  --name code-server --restart=on-failure:10   -p 127.0.0.1:8002:8080 \
  -v "$HOME/.config:/home/coder/.config" \
  -v "$HOME/webcode:/home/coder/code" \
  -u "$(id -u):$(id -g)" \
  -e "DOCKER_USER=$USER" \
  codercom/code-server:latest