
## Installation Instructions
1. Clone the repo into `/opt/port_mon`

```bash
cd /opt
git clone https://github.com/EricCavCore/port_mon.git
cd port_mon
```

2. Configure the `monitor.yml` file to target the IP and PORTs (with protocol specified)

`address:port/protocol`

3. Install the systemd service

```bash
cp port_mon.service /etc/systemd/system

systemctl daemon-reload
```

4. Start / enable the service
```bash
systemctl start port_mon
systemctl enable port_mon
```

5. Get logs from the file specified, default is `/opt/port_mon/port_mon.log`

```bash
cat port_mon.log
```
