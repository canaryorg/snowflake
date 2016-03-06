#!/bin/bash -eu

#--------------------------------------------------------
# install golang
#
sudo apt-get update
sudo apt-get install -y wget git
sudo wget -q -O /tmp/golang.tar.gz https://storage.googleapis.com/golang/go1.6.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf /tmp/golang.tar.gz

#--------------------------------------------------------
# install snowflake
#
cat <<EOF | sudo tee -a /etc/profile

export GOPATH="/home/ubuntu"
EOF

export GOPATH="/home/ubuntu"
/usr/local/go/bin/go get github.com/savaki/snowflake/cmd/snowflake

#--------------------------------------------------------
# create startup script
#
sudo rm -f /etc/rc.local
cat <<EOF | sudo tee -a /etc/rc.local
#!/bin/sh

nohup /home/ubuntu/bin/snowflake --port 80 & disown

EOF
sudo chmod +x /etc/rc.local
