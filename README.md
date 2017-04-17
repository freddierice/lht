# Linux Hacking Toolkit

A toolkit for managing multiple versions of linux, verifying, and testing exploits.

## Build
To get the project just
```bash
go get gopkg.in/freddierice/lht.v1/...
```
. This will download all of the dependencies and install the `lht` binary into `$GOPATH/bin`. Run
```bash
mv $GOPATH/bin/lht.v1 $GOPATH/bin/lht
sudo lht
```
for first time configuration. 

## Run
Lets say we want to start a project to start hacking on a variety of linux versions using arm. We could start the project with the following command:
```bash
lht project add proj1 -a arm -d ~/vexpress_defconfig -t arm-linux-gnueabihf
```

Now we want to build linux 4.9.7. To remember that this is the vulnerable version of linux for a given exploit, let's name it vuln. in this case, we would run
```bash
lht linux add proj1 4.9.7 -n vuln
```

Now let's download and build it:
```bash
lht linux build proj1 vuln
```
. Then, lht will start downloading and building linux, glibc, busybox and vuln-ko. Once it has completed, we can build a root filesystem: 
```bash
sudo lht fs create proj1 vuln
```
Note that we had to use `sudo` here. In the future this will not be necesary, but for the time being we need it to mount the root filesystem image.

## Vagrant
Optionally, a Vagrantfile is also available. To build with vagrant, go to `$GOPATH/src/gopkg.in/freddierice/lht.v1` and run `vagrant up`. This will download ubuntu with some cross compilation tools, go, and lht.

## TODO
 - Remove the need of ever needing root. Unfortunately, the only clean way to create a filesystem is to mount a raw file using loopback devices. Unfortunately this needs either root, or the capabililty `CAP_SYS_ADMIN`, which is essentially root.
 - Document `lht fs update`
