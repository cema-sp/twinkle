# -*- mode: ruby -*-
# vi: set ft=ruby :

# All Vagrant configuration is done below. The "2" in Vagrant.configure
# configures the configuration version (we support older styles for
# backwards compatibility). Please don't change it unless you know what
# you're doing.
Vagrant.configure(2) do |config|
  # The most common configuration options are documented and commented below.
  # For a complete reference, please see the online documentation at
  # https://docs.vagrantup.com.

  # Every Vagrant development environment requires a box. You can search for
  # boxes at https://atlas.hashicorp.com/search.
  # config.vm.box = "ubuntu/trusty64"

  # Disable automatic box update checking. If you disable this, then
  # boxes will only be checked for updates when the user runs
  # `vagrant box outdated`. This is not recommended.
  # config.vm.box_check_update = false

  # Create a forwarded port mapping which allows access to a specific port
  # within the machine from a port on the host machine. In the example below,
  # accessing "localhost:8080" will access port 80 on the guest machine.
  # config.vm.network "forwarded_port", guest: 27017, host: 27017

  # Create a private network, which allows host-only access to the machine
  # using a specific IP.
  # config.vm.network "private_network", ip: "192.168.33.10"

  # Create a public network, which generally matched to bridged network.
  # Bridged networks make the machine appear as another physical device on
  # your network.
  # config.vm.network "public_network"

  # Share an additional folder to the guest VM. The first argument is
  # the path on the host to the actual folder. The second argument is
  # the path on the guest to mount the folder. And the optional third
  # argument is a set of non-required options.
  # config.vm.synced_folder "../data", "/vagrant_data"

  # Provider-specific configuration so you can fine-tune various
  # backing providers for Vagrant. These expose provider-specific options.
  # Example for VirtualBox:
  #
  # config.vm.provider "virtualbox" do |vb|
  #   # Display the VirtualBox GUI when booting the machine
  #   vb.gui = true
  #
  #   # Customize the amount of memory on the VM:
  #   vb.memory = "1024"
  # end
  #
  # View the documentation for the provider you are using for more
  # information on available options.

  # Define a Vagrant Push strategy for pushing to Atlas. Other push strategies
  # such as FTP and Heroku are also available. See the documentation at
  # https://docs.vagrantup.com/v2/push/atlas.html for more information.
  # config.push.define "atlas" do |push|
  #   push.app = "YOUR_ATLAS_USERNAME/YOUR_APPLICATION_NAME"
  # end

  # Enable provisioning with a shell script. Additional provisioners such as
  # Puppet, Chef, Ansible, Salt, and Docker are also available. Please see the
  # documentation for more information about their specific syntax and use.

  config.vm.define "mongo" do |mongo|
    mongo.vm.box = "ubuntu/trusty64"
    mongo.vm.network "private_network", ip: "192.168.33.10"
    mongo.vm.provision "shell", inline: <<-SHELL
      sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv 7F0CEB10
      echo "deb http://repo.mongodb.org/apt/ubuntu "$(lsb_release -sc)"/mongodb-org/3.0 multiverse" | sudo tee /etc/apt/sources.list.d/mongodb-org-3.0.list
      sudo apt-get update
      sudo locale-gen
      # export LC_ALL=C
      sudo apt-get install -y mongodb-org
      sudo sed -i 's/bind_ip/#bind_ip/' /etc/mongod.conf
      sudo sed -i 's/#verbose/verbose/' /etc/mongod.conf
      sudo service mongod restart
    SHELL
  end

  config.vm.define "go" do |go|
    go.vm.box = "ubuntu/trusty64"
    go.vm.network "private_network", ip: "192.168.33.11"
    go.vm.synced_folder ".", "/home/vagrant/go/src/github.com/cema-sp/twinkle"
    go.vm.provision "shell", inline: <<-SHELL
      wget -q https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz -O - | tar -C /usr/local -xz
      echo "export PATH=$PATH:/usr/local/go/bin" | sudo tee -a /etc/profile
      echo "export GOROOT=/usr/local/go/" | sudo tee -a /etc/profile
      echo "export GOPATH=/home/vagrant/go/" | sudo tee -a /etc/profile
      source /etc/profile
      sudo apt-get update
      sudo locale-gen
      sudo apt-get install -y imagemagick libmagickwand-dev git
      pkg-config --cflags --libs MagickWand
      mkdir -p /home/vagrant/go/src/github.com/cema-sp/
      cd /home/vagrant/go/src/github.com/cema-sp/twinkle
      go get -t ./...
    SHELL
  end

end
