## These should be flagged
# <expect-error>
RUN sudo apt-get update && apt-get install -y vim

# <expect-error>
RUN sudo -u username yum install -y wget

# <expect-error>
RUN sudo dd if=/dev/zero of=/dev/sda

## These should not be flagged
RUN apt-get update && apt-get install -y vim
RUN yum install -y wget
