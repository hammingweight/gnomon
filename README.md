# gnomon

**gnomon** is a tool that monitors a SunSynk<sup>:registered:</sup> hybrid inverter and updates the inverter's settings based on the battery state of charge and input power. **gnomon** is
intended for small installations where:
 * The inverter's input (e.g. solar panels) and battery may not be sufficient to provide uninterrupted power without the grid, and
 * Loads are split into essential and non-essential loads

**gnomon** should run once a day, from sunrise to sunset. At the end of the running time, **gnomon** sets the maximum depth of discharge for the battery. Typically, a greater discharge depth will be configured in summer than in winter. The tool's goal is to maximize the amount of solar energy that is generated while ensuring that the battery has sufficient charge to handle power outages.

The software can also choose whether to power non-essential loads from the inverter (the inputs and battery) or from the grid.  **gnomon**'s heuristics use
on the battery's state of charge and how much power is being supplied by the inputs. In practice, **gnomon** will primarily 
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
```

