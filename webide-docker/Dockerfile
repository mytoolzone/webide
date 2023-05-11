FROM codercom/code-server:4.11.0-debian

#RUN curl -L https://go.dev/dl/go1.20.2.linux-amd64.tar.gz -o go.1.20.tar.gz 
WORKDIR /home/coder

RUN sudo apt-get update
RUN sudo apt-get install -y --no-install-recommends g++ gcc libc6-dev make pkg-config xz-utils;

## install goland
COPY go1.20.2.linux-amd64.tar.gz  ./
RUN sudo tar -C /usr/local -xzvf  go1.20.2.linux-amd64.tar.gz && rm  go1.20.2.linux-amd64.tar.gz
RUN sudo ln -s /usr/local/go/bin/go  /usr/local/bin 
RUN sudo ln -s /usr/local/go/bin/gofmt  /usr/local/bin 

## install node
COPY node-v18.15.0-linux-x64.tar.xz ./
RUN sudo tar -C /usr/local -xvf node-v18.15.0-linux-x64.tar.xz 
RUN rm node-v18.15.0-linux-x64.tar.xz
RUN sudo mv /usr/local/node-v18.15.0-linux-x64 /usr/local/node
RUN sudo ln -s /usr/local/node/bin/node  /usr/local/bin 

RUN sudo /usr/local/node/bin/npm config set registry https://registry.npm.taobao.org
RUN sudo /usr/local/node/bin/npm install -g pnpm && /usr/local/node/bin/npm install -g vue-cli

RUN code-server --install-extension golang.Go
RUN code-server --install-extension ms-python.python
## install mytool 

CMD [ "/bin/sh" ]


