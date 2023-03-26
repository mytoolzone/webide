FROM codercom/code-server:4.11.0-debian

#RUN curl -L https://go.dev/dl/go1.20.2.linux-amd64.tar.gz -o go.1.20.tar.gz 
WORKDIR /home/coder

## install goland
COPY go1.20.2.linux-amd64.tar.gz node-v18.15.0-linux-x64.tar.xz ./
RUN sudo tar -C /usr/local -xzvf  go1.20.2.linux-amd64.tar.gz && rm -f go1.20.2.linux-amd64.tar.gz
ENV PATH=$PATH:/usr/local/go/bin

RUN sudo apt-get update
RUN sudo apt-get install -y --no-install-recommends \
		g++ \
		gcc \
		libc6-dev \
		make \
		pkg-config \apt-get update;

COPY go1.20.2.linux-amd64.tar.gz node-v18.15.0-linux-x64.tar.xz ./


CMD [ "/bin/sh" ]




