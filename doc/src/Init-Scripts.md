Some parameters may require your attention, such as path and executable name.

## systemd
Create ```/etc/systemd/services/leng.service``` and paste in the following.

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

## init.d
Create ```/etc/init.d/leng``` and paste in the following.

```bash
#!/bin/bash
# leng daemon
# chkconfig: 345 20 80
# description: leng daemon
# processname: leng

DAEMON_PATH="/root/grim"

DAEMON=leng
DAEMONOPTS="-update"

NAME=leng
DESC="https://github.com/looterz/leng"
PIDFILE=/var/run/$NAME.pid
SCRIPTNAME=/etc/init.d/$NAME

case "$1" in
start)
	printf "%-50s" "Starting $NAME..."
	cd $DAEMON_PATH
	PID=`$DAEMON $DAEMONOPTS > /dev/null 2>&1 & echo $!`
	#echo "Saving PID" $PID " to " $PIDFILE
        if [ -z $PID ]; then
            printf "%s\n" "Fail"
        else
            echo $PID > $PIDFILE
            printf "%s\n" "Ok"
        fi
;;
status)
        printf "%-50s" "Checking $NAME..."
        if [ -f $PIDFILE ]; then
            PID=`cat $PIDFILE`
            if [ -z "`ps axf | grep ${PID} | grep -v grep`" ]; then
                printf "%s\n" "Process dead but pidfile exists"
            else
                echo "Running"
            fi
        else
            printf "%s\n" "Service not running"
        fi
;;
stop)
        printf "%-50s" "Stopping $NAME"
            PID=`cat $PIDFILE`
            cd $DAEMON_PATH
        if [ -f $PIDFILE ]; then
            kill -HUP $PID
            printf "%s\n" "Ok"
            rm -f $PIDFILE
        else
            printf "%s\n" "pidfile not found"
        fi
;;

restart)
  	$0 stop
  	$0 start
;;

*)
        echo "Usage: $0 {status|start|stop|restart}"
        exit 1
esac
```

## rc.d
Create ```/etc/rc.d/leng``` and paste in the following.

```sh
#!/bin/sh
#
# $FreeBSD$
#
# PROVIDE: leng
# REQUIRE: NETWORKING SYSLOG
# KEYWORD: shutdown
#
# Add the following lines to /etc/rc.conf to enable leng:
#
#leng_enable="YES"

. /etc/rc.subr

name="leng"
rcvar="leng_enable"

load_rc_config $name

: ${leng_user:="root"}
: ${leng_enable:="NO"}
: ${leng_directory:="/root/grim"}

command="${leng_directory}/leng -update"

pidfile="${leng_directory}/${name}.pid"

start_cmd="export USER=${leng_user}; export HOME=${leng_directory}; /usr/sbin/daemon -f -u ${leng_user} -p ${pidfile} $command"

#stop_cmd="kill $(cat $pidfile)"
stop_cmd="${name}_stop"
leng_stop() {
	if [ ! -f $pidfile ]; then
		echo "leng PID File not found. Maybe leng is not running?"
	else
		kill $(cat $pidfile)
	fi
}

run_rc_command "$1"
```
or
```#!/bin/sh
#
# OpenBSD
#
daemon="<path_to_daemon>"  

. /etc/rc.d/rc.subr

rc_bg=YES
rc_reload=NO

rc_cmd $1
```