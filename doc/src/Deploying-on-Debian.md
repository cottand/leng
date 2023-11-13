# Deploying on Debian

## Installing leng

Installing leng is the easiest when you simply download a release from the [releases](https://github.com/looterz/leng/releases) page. Go ahead and copy the link for leng_linux_x64 and run the following in your terminal.

```
mkdir ~/grim
cd ~/grim
wget <leng release>
```

This will download the binary to ```~/grim``` which will be leng's working directory. First, lets setup file permissions for leng, by running the following.

```
chmod a+x ./leng_linux_x64
```

Setup is pretty much complete, the only thing left to do is run leng and let it generate the default configuration and download the blocklists, but lets set it up as a systemd service so it automatically restarts and updates when starting.

## Setting up the service

Create the leng service by running the following,

```
nano /etc/systemd/system/leng.service
```

Now paste in the code for the service below,

```
[Unit]
Description=leng dns proxy
Documentation=https://github.com/looterz/leng
After=network.target

[Service]
User=root
WorkingDirectory=/root/grim
LimitNOFILE=4096
PIDFile=/var/run/leng/leng.pid
ExecStart=/root/grim/leng_linux_x64 -update
Restart=always
StartLimitInterval=30

[Install]
WantedBy=multi-user.target
```

Save, and now you can start, stop, restart and run status commands on the leng service like follows
```
systemctl start leng  # start
systemctl enable leng # start on boot
```

The only thing left to do is setup your clients to use your leng dns server.

## Security

Now that leng is setup on your droplet, it's recommended to [secure](https://github.com/looterz/leng/wiki/Securing-on-linux) the installation from non-whitelisted clients.