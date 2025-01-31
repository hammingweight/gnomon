# gnomon

**gnomon** is a tool that monitors a SunSynk hybrid inverter and updates the inverter's settings based on the battery state of charge and input power. **gnomon** is
intended for small installations where:
 * The inverter's input (e.g. solar panels) and battery may not be sufficient to provide uninterrupted power without the grid
 * Loads are split into essential and non-essential loads

**gnomon** is typically run once a day from sunrise to sunset. At the end of the running time, it sets the maximum depth of discharge for the battery. Typically, a greater discharge depth will be configured in summer than in winter. The tool's goal is to maximize the amount of solar energy that is generated while ensuring that the battery has sufficient charge to handle power outages.

The software can also be run to set whether non-essential loads should be powered by the inverter (from the inputs and battery) or from the grid. The heuristics used
 to make the decision are based on the battery's state of charge and how much power is being supplied by the inputs. In practice, **gnomon** will choose primarily to 
 use the grid to power non-essential loads in winter or overcast days. In summer, **gnomon** will try to balance using less grid power while maintaining a reasonable
battery state of charge.
