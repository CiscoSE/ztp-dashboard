# ztp-dashboard

Dashboard to manage zero touch provisioning configurations and automated upgrades for XR and NX devices


## Business/Technical Challenge

The on-boarding of devices into the network can be challenging. It can require manual configuration, and that increases the risk of human error.
Upgrading to newer software images can also be quite complex, and testing that everything work as expected is rarely automated.

Zero Touch Provisioning (ZTP) for XR and Power On Auto-Provisioning (POAP) for Nexus, can make this easier. Since these processes automate the 
on-boarding of devices, you are able to do the initial software install and the day-0 configuration without human intervention.

However, configuring ZTP and POAP in your environment requires knowledge around DHCP, HTTP and other tools. Also, if you want to do upgrades for 
devices already present in the network, you still need to manually save the configuration of the device and do the reboot with the correct options.

Finally, ZTP and POAP do not include automated tests.

## Proposed Solution

In order to enable customers to fully take advantage of ZTP and POAP, an application can be done where the following tasks are automated:

* Setup DHCP server configuration, including options and client identifiers
* Setup HTTP configuration, where XR and NX images will be stored along with day 0 scripts
* Detection of the different phases for the ZTP and POAP processes 
* Tests (ping or telemetry)
* Notifications 

The solution will help operators to configure HTTP, DHCP or TFTP from a single portal, without the need of extensive knowledge around how these technologies work. Since everything is managed from a single point, alerts with extensive descriptions 
can be sent to monitoring tools when troubleshooting needs to be done.

### Cisco Products Technologies/ Services

The solution will leverage the following Cisco technologies

* Cisco IOS XR
* Cisco Nexus
* Crosswork Situation Manager
* Webex Teams

## Team Members

* Don Green <dongree@cisco.com> - Americas Services Providers
* Jason Mah <jamah@cisco.com> - Americas Global Virtual Engineering
* Santiago Flores Kanter <sfloresk@cisco.com> - Americas Service Providers 

## Solution Components

* Golang

## Usage


## Installation


## Documentation


## License

Provided under Cisco Sample Code License, for details see [LICENSE](./LICENSE.md)

## Code of Conduct

Our code of conduct is available [here](./CODE_OF_CONDUCT.md)

## Contributing

See our contributing guidelines [here](./CONTRIBUTING.md)
