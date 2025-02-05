# gnomon
<img src="/images/gnomon.jpg" align="right" width="200px">

**gnomon** is a tool that monitors a SunSynk<sup>:registered:</sup> hybrid inverter and updates the inverter's settings based on the battery state of charge and input power. **gnomon** is
intended for small installations where:
 * The inverter's input (e.g. solar panels) and battery may not be sufficient to provide uninterrupted power without the grid, and
 * Loads are split into essential and non-essential loads

**gnomon** should run once a day, from sunrise to sunset. At the end of the running time, **gnomon** sets the maximum depth of discharge for the battery. Typically, a greater discharge depth will be configured in summer than in winter. The tool's goal is to maximize the amount of solar energy that is generated while ensuring that the battery has sufficient charge to handle power outages.

The software can also choose whether to power non-essential loads from the inverter (the inputs and battery) or from the grid.  **gnomon**'s heuristics use
the battery's state of charge and how much power is being supplied by the inputs. In practice, **gnomon** will primarily 
use the grid to power non-essential loads in winter or overcast days. On long, clear summer days, **gnomon** will try to use less grid power while maintaining a reasonable
battery state of charge.

## Installing and configuring *gnomon*
You can download **gnomon** for Windows, Mac or Linux from [releases](https://github.com/hammingweight/gnomon/releases).

**gnomon** uses a configuration file with credentials for authentication with the SunSynk API. The default location for the config file is
`$HOME/.synk/config`. The configuration file is a YAML file with a username, user password and an inverter serial number, like

```
$ cat $HOME/.synk/config
user: carl@example.com
password: "VerySecret"
default_inverter_sn: 2401010123
```

You can create the configuration file using an editor. Alternatively, you can create the file using [synkctl](https://github.com/hammingweight/synkctl) and running

```
$ synkctl configuration generate -u carl@example.com -p "VerySecret" -i 2401010123
```


## Running *gnomon*
Running **gnomon** with a `--help` flag shows the options

```
$ gnomon --help
gnomon is a tool for automatically managing a SunSynk inverter's settings.
It adjusts the depth of discharge of the battery and, optionally, can decide when to
allow the inverter to power non-essential loads.

Usage:
  gnomon [flags]

Flags:
  -c, --config string    synkctl config file path (default "/home/carl/.synk/config")
  -C, --ct-coil          manage power to the non-essential load
  -e, --end HH:MM        end time in 24 hour HH:MM format, e.g. 19:30
  -h, --help             help for gnomon
  -l, --logfile string   log file path
  -m, --min-soc SoC      minimum battery state of charge
  -s, --start HH:MM      start time in 24 hour HH:MM format, e.g. 06:00
  -v, --version          version for gnomon
```

For example, to run the script so that it starts managing the inverter at 5:00AM, stops managing at 7:30PM, overrides the default configuration file ("myconfig"), logs to a file "gnomon.log", won't allow the battery state of charge to drop below 40%, and manages
the CT coil to control whether the grid or inverter should power the non-essential loads, run **gnomon** with the following flags

```
gnomon -s 05:00 -e 19:30 -c myconfig -l gnomon.log -m 40 -C
```

All flags are optional and simply running

```
$ gnomon
```
will:
* start managing the inverter immediately (i.e. a start time of 'now')
* end management of the inverter in 12 hours
* use the default configuration file
* write logs to stdout
* set the minimum battery SoC to the low battery SoC plus 20%
* not manage power to the non-essential loads

The following is a snippet of the first few lines logged by **gnomon** when managing power to the non-essential load (via the CT coil)

```
$ gnomon -C
2025/01/31 16:23:28 Starting management of the inverter
2025/01/31 16:23:28 Starting power management to the CT
2025/01/31 16:23:28 Starting management of the battery SOC
2025/01/31 16:23:29 Input power = 1125W, Battery SOC = 84%, Load = 88W
2025/01/31 16:23:29 Minimum battery SoC = 40%
```

### Running *gnomon* as a cron job
While you can run **gnomon** manually, it's a better idea to run it daily using `cron` or as a Kubernetes `CronJob`. For example, 
with this as a `crontab` entry to run **gnomon** starting at 6:00AM (and ending at 8:00PM/20:00)

```
00 06 * * * gnomon -C -e 20:00 -l /home/carl/gnomon.logs
```

## Important Note: Permissions
If the logs show that updating the inverter settings failed with messages like

```
2025/01/31 21:00:00 Setting battery depth of discharge to 80%
2025/01/31 21:00:01 updating battery capacity failed:  No Permissions
```

you need to upgrade your SunSynk<sup>:registered:</sup> account from end-user to installer by completing an [online form submission](https://www.sunsynk.org/remote-monitoring).
