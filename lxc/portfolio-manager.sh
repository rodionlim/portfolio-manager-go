#!/usr/bin/env bash
source <(curl -s https://raw.githubusercontent.com/rodionlim/portfolio-manager-go/main/lxc/build.func)
# Copyright (c) 2025-2025 rodion
# Author: rodion (rodionlim)
# License: MIT | https://github.com/rodionlim/portfolio-manager-go/raw/main/LICENSE

APP="Portfolio-Manager"
var_tags="misc"
var_cpu="1"
var_ram="512"
var_disk="5"
var_os="debian"
var_version="12"
var_unprivileged="1"

header_info "$APP"
variables
color
catch_errors

function update_script() {
    header_info
    check_container_storage
    check_container_resources
    if [[ ! -d /opt/PortfolioManager ]]; then
        msg_error "No ${APP} Installation Found!"
        exit
    fi
    if (( $(df /boot | awk 'NR==2{gsub("%","",$5); print $5}') > 80 )); then
        read -r -p "Warning: Storage is dangerously low, continue anyway? <y/N> " prompt
        [[ ${prompt,,} =~ ^(y|yes)$ ]] || exit
    fi
    msg_info "Stopping ${APP}"
    systemctl stop PortfolioManager.service
    msg_ok "Stopped ${APP}"

    msg_info "Updating ${APP}"
    VERSION=$(curl -s https://raw.githubusercontent.com/rodionlim/portfolio-manager-go/main/VERSION)
    curl -L -o /opt/PortfolioManager/portfolio-manager https://github.com/rodionlim/portfolio-manager-go/releases/download/v${VERSION}/portfolio-manager_linux_amd64
    chmod +x /opt/PortfolioManager/portfolio-manager
    msg_ok "Updated ${APP} v${VERSION}"

    msg_info "Starting ${APP}"
    systemctl start PortfolioManager.service
    msg_ok "Started ${APP}"

    msg_ok "Updated Successfully"
    exit
}

start
build_container
description

msg_ok "Completed Successfully!\n"
echo -e "${CREATING}${GN}${APP} setup has been successfully initialized!${CL}"
echo -e "${INFO}${YW} Access it using the following URL:${CL}"
echo -e "${TAB}${GATEWAY}${BGN}http://${IP}:8080${CL}"
