Name: dms
Version: %{commit}
Release: %{os_version}
Summary: Actiontech %{name}
Source0: %{name}.tar.gz
License: Enterprise
Prefix: /usr/local/%{name}

%define debug_package %{nil}

%description
Acitontech %{name}

%prep
# unpack source quiet
%setup -q

%build
# build is done in outside, please see Makefile.

%install
rm -rf $RPM_BUILD_ROOT/usr/local/%{name}
mkdir -p $RPM_BUILD_ROOT/usr/local/%{name}
mkdir -p $RPM_BUILD_ROOT/usr/local/%{name}/service-file-template
mkdir -p $RPM_BUILD_ROOT/usr/local/%{name}/static/logo
cp %{_builddir}/%{buildsubdir}/%{name}/build/service-file-template/* $RPM_BUILD_ROOT/usr/local/%{name}/service-file-template/
cp %{_builddir}/%{buildsubdir}/%{name}/build/logo/* $RPM_BUILD_ROOT/usr/local/%{name}/static/logo/
cp %{_builddir}/%{buildsubdir}/%{name}/config.yaml $RPM_BUILD_ROOT/usr/local/%{name}/config.yaml
cp %{_builddir}/%{buildsubdir}/%{name}/database_driver_option.yaml $RPM_BUILD_ROOT/usr/local/%{name}/database_driver_option.yaml
cp -r %{_builddir}/%{buildsubdir}/%{name}/build/static/* $RPM_BUILD_ROOT/usr/local/%{name}/static/
cp %{_builddir}/%{buildsubdir}/%{name}/build/scripts/init_start.sh $RPM_BUILD_ROOT/usr/local/%{name}/init_start.sh
cp %{_builddir}/%{buildsubdir}/%{name}/bin/dms $RPM_BUILD_ROOT/usr/local/%{name}/dms

%clean
rm -rf %{_builddir}/%{buildsubdir}
rm -rf $RPM_BUILD_ROOT


%pre

#create group & user
((which nologin 1>/dev/null 2>&1) || (echo "require nologin" && exit 1)) || exit 11
(getent group %{group_name} 1>/dev/null 2>&1) || groupadd -g 5800 %{group_name}
(id  %{user_name} 1>/dev/null 2>&1) || useradd -M -g %{group_name} -s $(which nologin) -u 5800 %{user_name}

#create ulimit
if [ ! -d "/etc/security/limits.d" ];then
    mkdir /etc/security/limits.d
    chmod 0755 /etc/security/limits.d
fi

cat > /etc/security/limits.d/%{name}.conf <<EOF
%{user_name}     soft    nofile    65535
%{user_name}     hard    nofile    65535
%{user_name}     soft    nproc     65535
%{user_name}     hard    nproc     65535
EOF
chown root: /etc/security/limits.d/%{name}.conf
chmod 440 /etc/security/limits.d/%{name}.conf

# https://docs.fedoraproject.org/en-US/packaging-guidelines/Scriptlets/
%post


grep systemd /proc/1/comm 1>/dev/null 2>&1
if [ $? -eq 0 ]; then
    sed -e "s|PIDFile=|PIDFile=$RPM_INSTALL_PREFIX\/dms.pid|g" -e "s|ExecStart=|ExecStart=$RPM_INSTALL_PREFIX\/dms -conf $RPM_INSTALL_PREFIX/config.yaml|g" -e "s|WorkingDirectory=|WorkingDirectory=$RPM_INSTALL_PREFIX|g" $RPM_INSTALL_PREFIX/service-file-template/dms.systemd > /lib/systemd/system/dms.service

    systemctl daemon-reload
    systemctl enable dms.service
fi

function box_out()
{
  local s=("$@") b w
  for l in "${s[@]}"; do
    ((w<${#l})) && { b="$l"; w="${#l}"; }
  done
  tput setaf 3
  echo " -${b//?/-}-
| ${b//?/ } |"
  for l in "${s[@]}"; do
    printf '| %s%*s%s |\n' "$(tput setaf 2)" "-$w" "$l" "$(tput setaf 3)"
  done
  echo "| ${b//?/ } |
 -${b//?/-}-"
  tput sgr 0
}
box_out "To start the service, please run script ./init_start.sh in the $RPM_INSTALL_PREFIX directory" "Example: cd $RPM_INSTALL_PREFIX && ./init_start.sh"

# chown
chown -R %{user_name}: $RPM_INSTALL_PREFIX
# chmod
find $RPM_INSTALL_PREFIX -type d -exec chmod 0750 {} \;
find $RPM_INSTALL_PREFIX -type f -exec chmod 0640 {} \;
chmod 0750 $RPM_INSTALL_PREFIX/dms
chmod 0750 $RPM_INSTALL_PREFIX/init_start.sh

# set cap change run user
# setcap CAP_CHOWN,CAP_DAC_OVERRIDE,CAP_FOWNER,CAP_SETUID,CAP_SETGID+eip $RPM_INSTALL_PREFIX/dms

function kill_and_wait {
   pidfile=$1
   proc_name=$2
   main_proc=$3

   if [ -e $pidfile ]; then
       pid=$(cat $pidfile)
       kill -0 $pid &>/dev/null
       if [ $? -eq 0 ]; then
           if [ $main_proc = "$(cat /proc/$pid/comm)" ];then
               systemctl stop $proc_name.service &>/dev/null
               for i in {1..30}; do
                    if [ ! -e $pidfile ]; then
                       return 0
                    fi
                    kill -0 $pid &>/dev/null
                    if [ $? -ne 0 ]; then
                       return 0
                    fi
                    sleep 1
               done
           fi
       fi
   fi

   return 1
}

if [ "$1" = "2" ]; then
    kill_and_wait $RPM_INSTALL_PREFIX/dms.pid dms dms
    if [ $? -ne 0 ]; then
        (>&2 echo "kill dms failed  , maybe process has exited or kill dms timeout")
        (>&2 echo "if dms service still exists , please kill dms manually")
    else
        (>&2 echo "dms has been killed , please start dms service manually")
    fi
fi

%preun
function kill_and_wait {
   pidfile=$1
   proc_name=$2
   main_proc=$3

   if [ -e $pidfile ]; then
       pid=$(cat $pidfile)
       kill -0 $pid &>/dev/null
       if [ $? -eq 0 ]; then
           if [ $main_proc = "$(cat /proc/$pid/comm)" ];then
               systemctl stop $proc_name.service &>/dev/null
               for i in {1..30}; do
                    if [ ! -e $pidfile ]; then
                       return 0
                    fi
                    kill -0 $pid &>/dev/null
                    if [ $? -ne 0 ]; then
                       return 0
                    fi
                    sleep 1
               done
           fi
       fi
   fi

   return 1
}

if [ "$1" = "0" ]; then
    kill_and_wait $RPM_INSTALL_PREFIX/dms.pid dms dms
    if [ $? -ne 0 ]; then
        (>&2 echo "kill dms failed , maybe dms process has exited or kill dms timeout")
        (>&2 echo "if dms service still exists , please kill dms manually")
    else
        (>&2 echo "kill dms success")
    fi
fi



%postun
if [ "$1" = "0" ]; then
    rm -f /etc/security/limits.d/%{name}.conf

    grep systemd /proc/1/comm 1>/dev/null 2>&1
    if [ $? -eq 0 ]; then
        systemctl disable dms.service 
        rm -f /lib/systemd/system/dms.service 
        systemctl daemon-reload
        # 如果 service 处于 failed 状态时 , 重置状态
        systemctl reset-failed dms.service &>/dev/null || true
    fi
fi

%files
/usr/local/%{name}/service-file-template/*
/usr/local/%{name}/dms
/usr/local/%{name}/config.yaml
/usr/local/%{name}/database_driver_option.yaml
/usr/local/%{name}/init_start.sh
/usr/local/%{name}/static/* 
