#/bin/bash

function check() {
	$@
	if [ ! $? -eq 0 ]; then
		exit $?
	fi
}

check apt-get update -y
check apt-get upgrade -y
check apt-get build-dep -y linux-image-$(uname -r)
check apt-get install -y binutils-aarch64-linux-gnu binutils-arm-linux-gnueabihf binutils-multiarch binutils-arm-linux-gnueabi binutils-arm-none-eabi binutils-avr gcc-aarch64-linux-gnu gcc-arm-linux-gnueabihf gcc-arm-linux-gnueabi gcc-avr
check apt-get install -y git wget

check wget --quiet "https://storage.googleapis.com/golang/go1.8.1.linux-amd64.tar.gz"
check tar -C /usr/local -xf go1.8.1.linux-amd64.tar.gz
check rm -f go1.8.1.linux-amd64.tar.gz

check echo "export GOPATH=/opt/go" >> /home/ubuntu/.bashrc
check echo "export PATH=\$PATH:/usr/local/go/bin:/opt/go/bin" >> /home/ubuntu/.bashrc

export GOPATH=/opt/go
export PATH=$PATH:/usr/local/go/bin:/opt/go/bin
check go get gopkg.in/freddierice/lht.v1/...
check mv /opt/go/bin/lht.v1 /opt/go/bin/lht

export SUDO_UID=1000
export SUDO_GID=1000
/opt/go/bin/lht
