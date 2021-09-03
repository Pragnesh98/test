#Add Asterisk User
adduser asterisk
passwd asterisk
usermod -aG wheel asterisk

#Pre Menuselect
sudo yum -y install epel-release kernel-devel make gcc gcc-c++ pkgconfig wget bzip2 patch python-devel ncurses-devel libxml2-devel libubsan libuuid-devel jansson-devel sqlite sqlite-devel unixODBC-devel subversion libtool-ltdl-devel postgresql-odbc postgresql-devel vim unzip sendmail libedit-devel file openssl-devel
sudo wget http://thrysoee.dk/editline/libedit-20190324-3.1.tar.gz
tar xzvf libedit-20190324-3.1.tar.gz
cd libedit-20190324-3.1
sudo ./configure && sudo make && sudo make install
sudo setenforce 0
cd /usr/src/
sudo wget http://downloads.asterisk.org/pub/telephony/asterisk/asterisk-17-current.tar.gz
sudo tar xvfz asterisk-17*
cd asterisk-17*
sudo ./configure

#Menuselect
sudo make menuselect

#Post menuselect
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
mkdir -p /home/asterisk/asterisk-ari/
cd /home/asterisk/asterisk-ari/

#Git Clone
git clone

#Post Git
sudo cp /home/asterisk/asterisk-ari/utils/asterisk/sip.conf /etc/asterisk/
sudo cp /home/asterisk/asterisk-ari/utils/asterisk/internal/asterisk.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable asterisk
sudo systemctl daemon-reload
sudo systemctl start asterisk

export NEW_RELIC_LICENSE_KEY="34ff2b9dd449161b5e0a5bbc161c06877355NRAL"
sudo mkdir -p /var/log/yellowmessenger/asterisk_ari/
sudo touch /var/log/yellowmessenger/asterisk_ari/asterisk_ari.log
sudo chown -R asterisk:asterisk /var/log/yellowmessenger/
mkdir -p /home/asterisk/asterisk-ari/logs/
sudo cp /home/asterisk/asterisk-ari/utils/asterisk/asterisk-ari.service /etc/systemd/system/
sudo systemctl enable asterisk-ari
sudo systemctl start asterisk-ari