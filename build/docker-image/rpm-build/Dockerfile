FROM centos:7

# install rpmbuild
RUN yum install -y rpmdevtools yum-utils
ADD docker-rpm-build.sh /usr/local/bin
RUN chmod +x /usr/local/bin/docker-rpm-build.sh
RUN rpmdev-setuptree

# install golang
ADD https://mirrors.ustc.edu.cn/golang/go1.19.6.linux-amd64.tar.gz /usr/local
RUN cd /usr/local \
    && tar -zxvf /usr/local/go1.19.6.linux-amd64.tar.gz
RUN cp /usr/local/go/bin/* /bin

# install make
RUN yum install -y epel-release
RUN yum update -y
RUN yum install -y make

# intall gcc
RUN yum install -y gcc

# install wget
RUN yum install -y wget

# install git
RUN yum install -y git

# clean
RUN yum clean all && \
    rm -r -f /var/cache/*


ENTRYPOINT ["/usr/local/bin/docker-rpm-build.sh"]
