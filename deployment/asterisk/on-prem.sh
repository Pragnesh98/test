

#Add Asterisk User
adduser asterisk
passwd asterisk
usermod -aG wheel asterisk
sudo visudo

sudo yum -y install tmux

vim ~/.tmux.conf
vim /home/asterisk/.tmux.conf
```
bind M-Left select-pane -L
bind M-Right select-pane -R
bind M-Up select-pane -U
bind M-Down select-pane -D

set -g default-terminal "screen-256color"
# Alt-n for new window
bind C-b new-window

bind -n S-Left previous-window
bind -n S-Right next-window
```

tm() { tmux new -s "$1" ;}
ta() { tmux attach -t "$1"; }
tl() { tmux list-sessions; }

#Pre Menuselect
sudo yum -y install epel-release kernel-devel make gcc gcc-c++ pkgconfig wget bzip2 patch python-devel ncurses-devel libxml2-devel libubsan libuuid-devel jansson-devel sqlite sqlite-devel unixODBC-devel subversion libtool-ltdl-devel postgresql-odbc postgresql-devel vim unzip sendmail libedit-devel file openssl-devel git libtermcap-devel libuuid-devel libxml2-devel sqlite-devel alsa-lib-devel fftw-devel curl-devel
sudo wget http://thrysoee.dk/editline/libedit-20190324-3.1.tar.gz
tar xzvf libedit-20190324-3.1.tar.gz
cd libedit-20190324-3.1
sudo ./configure && sudo make && sudo make install
git clone https://kashish_r@bitbucket.org/yellowmessenger/asterisk-ari.git
sudo setenforce 0
cd /usr/src/
sudo wget http://downloads.asterisk.org/pub/telephony/asterisk/asterisk-17-current.tar.gz
sudo wget http://downloads.asterisk.org/pub/telephony/asterisk/old-releases/asterisk-17.9.0.tar.gz
sudo wget http://downloads.asterisk.org/pub/telephony/asterisk/old-releases/asterisk-17.1.0.tar.gz
sudo tar xvfz asterisk-17*
cd asterisk-17*
sudo ./configure --with-jansson-bundled

#Menuselect
sudo cp /home/asterisk/asterisk-ari/deployment/asterisk/menuselect.makeopts /usr/src/asterisk-17.1.0/
sudo make menuselect

#Post menuselect
cd /usr/src/asterisk-17*
sudo make
sudo contrib/scripts/get_mp3_source.sh
sudo make install
sudo make samples
sudo chown -R asterisk:asterisk /etc/asterisk/
sudo chown -R asterisk:asterisk /var/spool/asterisk/
sudo chown -R asterisk:asterisk /var/run/asterisk/
sudo chown -R asterisk:asterisk /var/lib/asterisk/
sudo chown -R asterisk:asterisk /var/log/asterisk/
sudo mv /etc/asterisk/sip.conf /etc/asterisk/sip.conf.sample
sudo mv /etc/asterisk/extensions.conf /etc/asterisk/extensions.conf.sample
#mkdir -p /home/asterisk/asterisk-ari/
cd /home/asterisk/asterisk-ari/

#Install Docker
sudo yum install -y yum-utils
sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
sudo yum install docker-ce docker-ce-cli containerd.io
sudo systemctl status docker
sudo systemctl start docker

#Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/1.29.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
sudo ln -s /usr/local/bin/docker-compose /usr/bin/docker-compose

#Install MySQL
sudo dnf install mysql-server
sudo systemctl start mysqld.service
sudo systemctl status mysqld
sudo mysql_secure_installation


#Git Clone
cd /home/asterisk/
git clone https://kashish_r@bitbucket.org/yellowmessenger/asterisk-ari.git

#Post Git
sudo cp /home/asterisk/asterisk-ari/deployment/asterisk/sip.conf /etc/asterisk/
cp /home/asterisk/asterisk-ari/deployment/asterisk/extensions.conf /etc/asterisk/
sudo cp /home/asterisk/asterisk-ari/deployment/asterisk/asterisk.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable asterisk
sudo systemctl daemon-reload
sudo systemctl start asterisk

export NEW_RELIC_LICENSE_KEY="34ff2b9dd449161b5e0a5bbc161c06877355NRAL"
sudo mkdir -p /var/log/yellowmessenger/asterisk_ari/
sudo touch /var/log/yellowmessenger/asterisk_ari/asterisk_ari.log
sudo chown -R asterisk:asterisk /var/log/yellowmessenger/
mkdir -p /home/asterisk/asterisk-ari/logs/
sudo cp /home/asterisk/asterisk-ari/deployment/asterisk/asterisk-ari.service /etc/systemd/system/
sudo chown -R asterisk:asterisk /home/asterisk/asterisk-ari
mkdir /home/asterisk/.google/
sudo systemctl enable asterisk-ari
sudo systemctl start asterisk-ari

#Installing New Relic Infrastructure agent: https://one.newrelic.com/launcher/nr1-core.settings?platform[accountId]=2546315&pane=eyJuZXJkbGV0SWQiOiJzZXR1cC1uZXJkbGV0LnNldHVwLW9zIiwiZGF0YVNvdXJjZSI6IkNFTlRPUyJ9
# Create a configuration file and add your license key \
echo "license_key: 34ff2b9dd449161b5e0a5bbc161c06877355NRAL" | sudo tee -a /etc/newrelic-infra.yml && \
\
# Create the agent’s yum repository \
sudo curl -o /etc/yum.repos.d/newrelic-infra.repo https://download.newrelic.com/infrastructure_agent/linux/yum/el/8/x86_64/newrelic-infra.repo && \
\
# Update your yum cache \
sudo yum -q makecache -y --disablerepo='*' --enablerepo='newrelic-infra' && \
\
# Run the installation script \
sudo yum install newrelic-infra -y

#sngrep
sudo vim /etc/yum.repos.d/irontec.repo
[irontec]
name=Irontec RPMs repository
baseurl=http://packages.irontec.com/centos/8/$basearch/

sudo rpm --import http://packages.irontec.com/public.key
sudo yum -y install sngrep

#FFMPEG
sudo yum install https://dl.fedoraproject.org/pub/epel/epel-release-latest-8.noarch.rpm
sudo yum install https://download1.rpmfusion.org/free/el/rpmfusion-free-release-8.noarch.rpm https://download1.rpmfusion.org/nonfree/el/rpmfusion-nonfree-release-8.noarch.rpm
sudo yum install http://rpmfind.net/linux/epel/7/x86_64/Packages/s/SDL2-2.0.14-2.el7.x86_64.rpm
sudo yum install ffmpeg ffmpeg-devel

#SoX
sudo dnf install sox
sudo cp /usr/bin/sox /usr/local/bin/
 - sudo cp /usr/bin/ffmpeg /usr/local/bin/