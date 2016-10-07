# Simulation Engine

**WORK IN PROGRESS**

It's a web based simulation and visualization engine.

## Visualization

The engine starts a simulation program and expects the program to write to
_stdout_ single-line JSON encoded commands (one JSON per line).
Each line is parsed as a command which mostly instruct the engine to update
properties of certain objects.
As the positioning properties (currently 2D only) are mandatory, the engine
runs a web server which visualize the objects according to the positioning
and other properties which is object specific.

The keyboard and mouse events are sent back to simulation program via _stdin_
and encoded in single-line JSON too. Each line indicates a certain event.

### Object Rendering

As positioning properties are common, the engine will take care of positioning
and zooming of the view area.
For object specific properties, the engine requires rendering extensions corresponding
to object class.
The engine provides a few built-in rendering extensions, and is also extendable
at runtime using `-I` option to include additional Javascript and CSS files.

## Get Started

**LINUX IS THE ONLY SUPPORTED PLATFORM**

Build the tool using [hmake](https://evo-cloud.github.io/hmake) is hassle-free,
no worry about dependencies:

```
hmake
```

It will generate executable `bin/sim-ng`.
And now you can hook up your own simulation program:

```
bin/sim-ng vis -- my-sim-prog args...
```

Point your browser to `http://localhost:3500` and you will see the objects emitted
from your simulation program.

To hook up your own rendering extensions:

```
bin/sim-ng vis -I ext-dir1 -I ext-dir2 ... -- my-sim-prog args...
```

In each of extension directory, file `visualizer.plugin` is expected.
It's a YAML file, with simple information like:

```yaml
---
name: my-ext
visualizer:
  stylesheets:
    - styles.css
  scripts:
    - objects.js
```

## Example/Reference

Project [ZuPi](https://github.com/evo-bots/zupi) is an example using this simulator.
