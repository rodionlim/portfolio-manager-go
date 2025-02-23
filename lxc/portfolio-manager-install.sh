#!/usr/bin/env bash

# Copyright (c) 2021-2025 rodionlim
# Author: rodion (rodionlim)
# License: MIT
# https://github.com/rodionlim/portfolio-manager-go/blob/main/LICENSE

source /dev/stdin <<< "$FUNCTIONS_FILE_PATH"
color
verb_ip6
catch_errors
setting_up_container
network_check
update_os

msg_info "Installing Dependencies"
$STD apt-get install -y curl
$STD apt-get install -y sudo
$STD apt-get install -y mc
msg_ok "Installed Dependencies"

msg_info "Installing Portfolio-Manager"
# Create target directory if it doesn't exist
$STD mkdir -p /opt/PortfolioManager
$STD curl -L -o /opt/PortfolioManager/portfolio-manager https://github.com/rodionlim/portfolio-manager-go/releases/download/v1.0.0/portfolio-manager_linux_amd64
$STD chmod +x /opt/PortfolioManager/portfolio-manager
msg_ok "Installed Portfolio-Manager"

msg_info "Creating Service"
cat <<EOF >/etc/systemd/system/PortfolioManager.service
[Unit]
Description=Portfolio-Manager: Tool to price and manage portfolio
ConditionFileIsExecutable=/opt/PortfolioManager/portfolio-manager
After=syslog.target network-online.target

[Service]
StartLimitInterval=5
StartLimitBurst=10
ExecStart=/opt/PortfolioManager/portfolio-manager
WorkingDirectory=/opt/PortfolioManager
StandardOutput=file:/var/log/PortfolioManager.out
StandardError=file:/var/log/PortfolioManager.err
Restart=always
RestartSec=10
EnvironmentFile=-/etc/sysconfig/PortfolioManager

[Install]
WantedBy=multi-user.target
EOF
systemctl enable -q --now PortfolioManager.service
msg_ok "Created Service"

motd_ssh
customize

msg_info "Cleaning up"
$STD apt-get -y autoremove
$STD apt-get -y autoclean
msg_ok "Cleaned"
