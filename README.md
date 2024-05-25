# golog

[![Tests](https://github.com/bbathe/golog/workflows/Tests/badge.svg)](https://github.com/bbathe/golog/actions) [![Release](https://github.com/bbathe/golog/workflows/Release/badge.svg)](https://github.com/bbathe/golog/actions)

Minimalistic logging application for [Amateur Radio](https://www.arrl.org).

## Description
There are really just 2 main features in this application: QSO logging and DX Spotting.  QSO data is captured either by manual entry or automatically pulled in from ADIF files. DX Spotting is provided via an integration with [HamAlert](https://hamalert.org).

A very minimal set of data is captured to record a QSO:
* Station callsign
* QSO partner callsign
* Band
* Mode
* Date
* Time
* RST received
* RST sent

The philosphy is that if you want to find additional information on your QSO partner (address, other hobbies, etc.) goto a site like [QRZ.com](https://www.qrz.com).  And, even though this application stores QSO data, the final resting place is really at sites like [Logbook of the World](https://lotw.arrl.org), [Club Log](https://clublog.org), and the [QRZ Logbook](https://logbook.qrz.com) which provide a longitudinal record of all your QSOs, aggregated across the various logging application you will use in your lifetime.

## Installation
To install this application:

1. Create the folder `C:\Program Files\golog`
2. Download the `golog.exe.zip` file from the [latest release](https://github.com/bbathe/golog/releases) and unzip it into that folder
3. Double-click on the `golog.exe` file to start the application and finish the configuration
4. Create a shortcut somewhere or pin to taskbar to make it easier to start in the future

You can have multiple configuration files and switch between them by using the `config` command line switch:
  ```yaml
  golog.exe -config fieldday.yaml
  ```