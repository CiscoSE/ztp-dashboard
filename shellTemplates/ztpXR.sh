#!/bin/bash

# Source XR Helpers
source /pkg/bin/ztp_helper.sh
source /pkg/etc/ztp.config

umask 022

ztp_console_log "INFO: Zero Touch Provisioning started"

# DNS 
dns_servers="8.8.8.8 8.8.4.4"

[ "${new_domain_name_servers}" ] && dns_servers="${new_domain_name_servers}"

function dns_config() {
     local address
     for address in ${dns_servers}; do
          echo "nameserver ${address}" >> /etc/resolv.conf
     done
     echo "search cisco.com" >> /etc/resolv.conf
}
dns_config

# Configuration

config_url="{{.ServerURL}}{{.ConfigURL}}"


function configure_crypto() {
	# don't regenerate if we already have a host key
	if [ -z "$(xrcmd 'show crypto key mypubkey rsa')" ]; then
		echo "2048" | xrcmd "crypto key generate rsa"
		ztp_console_log "CONFIG: crypto key configured"
	fi
}
configure_crypto


config_file="${ZTP_DIR}/customer/ztp.config"

ztp_console_log "CONFIG: Getting XR config from ${config_url}/..."

rc=

rm -f "${config_file}"
curl --silent --connect-timeout 10 --retry 5 \
	--fail --location --output "${config_file}" "${config_url}"
rc="$?"

if [ "${rc}" -ne 0 ]; then
	ztp_console_log "Failed to get config from ${config_url}: curl exit status: ${rc}"
	ztp_hook_error_exit "$*"
	break
fi

if [ ! -f "${config_file}" ]; then
	ztp_console_log "Failed to get config from ${config_urls}: file not found"
	ztp_hook_error_exit "$*"
fi

xr_apply_config() {
	local d=/pkg
	d=${d}/bin
	xrnns_exec ${d}/config -p15 -R -f "$1"
}

# Applying config
ztp_console_log "CONFIG: Applying XR config from ${url}"
xr_apply_config "${config_file}"
ztp_console_log "CONFIG: XR configuration loaded from ZTP"

ztp_console_log "INFO: Zero Touch Provisioning completed"

# Notify that device is read
curl -X PUT {{.ServerURL}}/api/devices/provisioned