# smartmeter

Connects to P1 DSMR interface.

Influenced by [smeterd](https://github.com/nrocco/smeterd).

Licensed under MIT license.

## Development

```
go install github.com/koesie10/smartmeter/cmd/smartmeter
```

## Running

### As systemd service

```
sudo mv smartmeter /usr/local/bin
sudo chmod +x /usr/local/bin/smartmeter

sudo adduser --system --no-create-home --group smartmeter
sudo usermod -a -G dialout smartmeter

sudo nano /etc/systemd/system/smartmeter.service
```

```
[Unit]
Description=smartmeter
Wants=network-online.target
After=network-online.target
After=influxdb.service
AssertFileIsExecutable=/usr/local/bin/smartmeter

[Service]
User=smartmeter
Group=smartmeter

PermissionsStartOnly=true

Restart=always

ExecStart=/usr/local/bin/smartmeter publish --influx-database telegraf --influx-tags="house=myhouse"

[Install]
WantedBy=multi-user.target
```

```
sudo systemctl daemon-reload
sudo systemctl start smartmeter
sudo systemctl status smartmeter
sudo systemctl enable smartmeter
```

## Todo

- Checksum validation. I only have DSMR 2, so it was not necessary for me to implement.
