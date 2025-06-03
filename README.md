# Healthchecks Agent (Go)

Simple agent to ping Healthchecks.io with local service check results.

hcw_0EyratoqQAHMhydS7vZzg6XEiT1B
API key (read-only) hcr_NCRKdVlnAAx8PSwvYAsVFxCNAGp3


<!-- CARA BUILD  -->
go build -o bin/d-agent-healthchecks ./cmd/agent
chmod +x build/build-deb.sh build/build-rpm.sh



mkdir -p dist/etc/systemd/system
mkdir -p dist/usr/local/bin
mkdir -p dist/etc/d-agent-healthchecks


chmod +x build/build-deb.sh build/build-rpm.sh
./build/build-deb.sh
./build/build-rpm.sh


sudo systemctl daemon-reload
sudo systemctl enable d-agent-healthchecks
sudo systemctl start d-agent-healthchecks
sudo systemctl status d-agent-healthchecks


# Debian/Ubuntu
wget https://github.com/devetek/d-agent-healthchecks/releases/download/v0.1.0/d-agent-healthchecks-0.1.0-amd64.deb
sudo dpkg -i d-agent-healthchecks-0.1.0-amd64.deb

# CentOS/Rocky
wget https://github.com/devetek/d-agent-healthchecks/releases/download/v0.1.0/d-agent-healthchecks-0.1.0-1.x86_64.rpm
sudo rpm -ivh d-agent-healthchecks-0.1.0-1.x86_64.rpm


cara gampang curl -sSL https://raw.githubusercontent.com/devetek/d-agent-healthchecks/main/install.sh | bash

git clone https://github.com/devetek/d-agent-healthchecks
cd d-agent-healthchecks
chmod +x install.sh
./install.sh


#REMOVE
sudo systemctl stop d-agent-healthchecks
sudo systemctl disable d-agent-healthchecks
sudo rm /etc/systemd/system/d-agent-healthchecks.service
sudo systemctl daemon-reload
sudo dpkg -r d-agent-healthchecks
sudo rm -rf /etc/d-agent-healthchecks/
sudo rm -f /usr/local/bin/d-agent-healthchecks
sudo rm -rf /usr/share/d-agent-healthchecks/
sudo rm -rf /var/lib/dpkg/info/d-agent-healthchecks.*



#INSTALL
go build -o build/bin/d-agent-healthchecks ./cmd/agent
./build/build-deb.sh
sudo dpkg -i d-agent-healthchecks_0.1.0_amd64.deb 
sudo systemctl daemon-reexec
sudo systemctl daemon-reload
sudo systemctl enable d-agent-healthchecks
sudo systemctl start d-agent-healthchecks
sudo systemctl status d-agent-healthchecks

sudo systemctl restart d-agent-healthchecks