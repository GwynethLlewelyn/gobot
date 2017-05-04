# gobot
## PhD work ported to Go

![botmover logo](templates/images/botmover-logo-70px.jpg)

One day, the whole thing will be described here.
For now, all you need to know is that this works only on [OpenSimulator](http://opensimulator.org).

## Configuration

- Install gobot as any other Go application (`go get github.com/GwynethLlewelyn/gobot` should do the trick)
- Create a database in SQLite3 using `database/schema.sql`
- Remember that path and change `config.toml` accordingly
- The directories `lib/` and `templates/` only have static content, so either you configure `config.toml` to point to the right directories (if running `gobot` as a standalone Go application) or you get these directories directly served by your webserver/reverse proxy/whatever
- The `lsl/` directory just holds LSL (Linden Scripting Language) scripts, which are to be placed inside in-world objects, and they will *not* be served by `gobot` (or the webserver/reverse proxy)

If you're placing Gobot behind a nginx server, [this is the configuration you'll need](nginx-config.md)

## License

This is currently copyrighted until I finish my PhD, then you can copy it at will. If you wish to use my code **now**, then please ask me before! Thank you!