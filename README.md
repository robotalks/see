# General Purpose Visualization Engine

This is a simple object based 2D visualization engine for visualizing objects
on a single web page.

## Philosophy

The engine watches on a message source for JSON based messages.
A message contains a list of objects with status.
When a message arrives, the engine memorizes all objects with status and
forwards the changes to web pages which connects to the engine via web socket.
The web page visualizes the objects using renders according to the object types.

The message source can be a pipe to a process, a subscription to MQTT topics, or
anything else. It is extensible.

In additional to messages, the web page also listens for interactive events
like mouse, keyboard, and forwards the events to message source regardless if
it accepts.

## Getting Started

Supported platforms:

- Mac OS X
- Linux AMD64/ARM64/ARMv7

Build using Go:

```sh
go build -o bin/see ./cmd/see
```

Build for specific target OS/Arch:

```sh
./scripts/build.sh linux-amd64
./scripts/build.sh darwin-arm64
```

It will generate executable `bin/OS/ARCH/see`.
And now you can hook up your own simulation program:

```
bin/see -- my-sim-prog args...
```

Point your browser to `http://localhost:3500` and you will see the objects emitted
from your simulation program.

To watch an MQTT topic:

```
bin/see mqtt://server:port/topic-prefix
```

And it will watch messages from topic `topic-prefix/msgs`, and emits events to
`topic-prefix/events`.

## Renders in Plugins

To hook up your own rendering extensions:

```
bin/see -I ext-dir1 -I ext-dir2 ... -- my-sim-prog args...
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

The following directories are always scanned for plugins before anything else:

- `$HOME/.robotalks`
- Current directory

## Details

### Messages to visualizer

#### Reset the world

```json
{
  "action": "reset"
}
```

#### Update/Insert an object

```json
{
  "action": "object",
  "object": {
    "id": "object-id",
    "type": "object-type",
    "style": "css-class",
    "styles": ["css-class"],
    "rect": { "x": 0, "y": 0, "w": 0, "h": 0 },
    "origin": { "x": 0, "y": 0 },
    "radius": 0
  }
}
```

Other properties are type specific.

##### Camera

```json
{
  "type": "camera",
  "angle": 0,
}
```

##### Corner

```json
{
  "type": "corner",
  "loc": "lt|rb|lb|rt"
}
```

##### Dot

```json
{
  "type": "dot"
}
```

##### Image

```json
{
  "type": "image",
  "src": "url"
}
```

Keyword `TIMESTAMP` in `src` will be replaced with current timestamp if present.

##### Label

```json
{
  "type": "label",
  "content": "text to show"
}
```

##### Joystick

```json
{
  "type": "joystick"
}
```

#### Remove an object

```json
{
  "action": "remove",
  "id": "object-id"
}
```

### Events from visualizer

```json
[
  {
    "action": "click",
    "position": { "x": 0, "y": 0 },
  },
  {
    "action": "keyup/keydown/keypress",
    "key": {
      "repeat": false,
      "charCode": 0,
      "code": "ArrowDown/KeyI/...",
      "key": "ArrowDown/i/...",
      "keyCode": 40, /* 73 for i */
      "ctrl": false,
      "alt": false,
      "shift": false,
      "meta": false,
    }
  }
]
```

## License
MIT

## Example/Reference

Project [ZuPi](https://github.com/evo-bots/zupi) is an example using this visualizer.
