# gobot
## PhD work ported to Go

![botmover logo](mstile-310x310.png)

`gobot` is a simplified prototype of a cluster of intelligent agents roaming a virtual world (theoretically either in Second Life or OpenSimulator). Users can visually programme the paths/activities such agents are supposed to follow by __visually__ dragging simple markers around the virtual world, without the need of learning any computer language. The principle is similar, and based on, several computer games where the player also gives visual commands to an AI controlling the players' units (like in Command & Conquer, StarCraft, or The Sims).

Right now, the Go application just deals with visualising the support database and adds some tools to give manual commands to the in-world markers (known simply as 'cubes' in `gobot`); a prototype run of the engine is being ported from previous code originally in PHP. Go, because of its very simple and intuitive approach to multithreading, will be used to allow each agent to get their own thread running the AI engine.
 
One day, the whole thing will be fully described here!
For now, all you need to know is that this works only on [OpenSimulator](http://opensimulator.org) since Second Life does not allow intelligent agents to be created and run in the same way as OpenSimulator does.

## Server-Side Configuration

- Install `gobot` as any other Go application (`go get github.com/GwynethLlewelyn/gobot` should do the trick)
- Create a database in MySQL using `database/schema.sql` ([see below](#notes) why SQLite is not recommended any more)
- Create a new user in MySQL with an email address and a MD5-hashed password (something like `insert into Users (Email, Password) Values ('valid@email.address', '4405c5984441a1b86bec717dc063ca46');`), you'll need at least one user to login to the backoffice; use `echo "password"|md5sum` or an online MD5 generator to get a valid password hash; afterwards, you can add more users manually
- Copy `config.toml.sample` to `config.toml` and change it — remember the installation path and you should also set an URL to grab a map tile from your OpenSimulator environment) as well as your own 4-digit PIN for `LSLSignaturePIN`
- In Un*x machines, `gobot` ought to be run either in the background or with `screen` (which will allow you to see the `stderr` console as well)
- Point your browser to the URL of the `gobot` appplication, login with the email/password, and try things out on the menus

### Notes
 - The directories `lib/` and `templates/` only have static content, so either you configure `config.toml` to  point to the right directories (if running `gobot` as a standalone Go application) or you get these directories directly served by your webserver/reverse proxy/whatever
 - There has been a move from `sqlite3` to MySQL; `gobot` does not support `sqlite3` any longer because it seems to have problems with locks when there are too many goroutines reading the database file

If you're placing `gobot` behind a nginx server, [this is the configuration you'll need](nginx-config.md). Note that Go is wonderful as it includes its own webserver, so running it behind a 'real' web server is not necessary, although a real web server should be able to provide things like caching and direct serving of static content (images, JS, CSS...) for the backoffice, to make it even faster.

`gobot` supports the [go-logging](https://github.com/op/go-logging) logging backend, which will redirect the logs simultaneously to `stderr`, a log file, and (by default) criticalhe  messages to `syslog` as well. The file-based log is rotated automatically with [lumberjack](https://gopkg.in/natefinch/lumberjack.v2). All logging options can be changed on the `config.toml` configuration file as well.

If you wish to run `gobot` as a systemd service under Ubuntu, then just do  

	cd /etc/systemd/system/multi-user.target.wants
	ln -s {Your Root Directory Where You Placed The Gobot Files}/gobot.service
	systemctl daemon-reload
  
From then on, you can just use `service gobot start|stop` to launch or stop `gobot` as any other service (yes, `journalctl -u gobot.service` will work as well); to disable it from starting on reboot, use `systemctl disable gobot.service` or remove the symlink. 

## Dependencies

These will be installed, *but* you *should* make sure you have the latest versions of them! (use git pull on the directories)

- [agGrid](https://www.ag-grid.com/)
- [Bootstrap](http://getbootstrap.com/) and [Bootstrap Dialog](https://nakupanda.github.io/bootstrap3-dialog/)
- StartBootstrap's [SB Admin 2](https://startbootstrap.com/template-overviews/sb-admin-2/) template
- [Leaflet.js](http://leafletjs.com/) (called remotely, no need to worry about it)
- [Leaflet.vector-markers](https://github.com/hiasinho/Leaflet.vector-markers) (to get cute markers for Leaflet)
- [Prism.js](http://prismjs.com/) (called remotely, no need to worry about it)
- Gravatar Hovercard support (also called remotely except for one file)

## Configuration on the Virtual World

Here is one of the toughest parts of the configuration. 

Move over to the _LSL scripts_ menu option. These should pre-generate correctly configured LSL scripts for the many in-world scripts. Copy & paste them inside the LSL editor on the SL Viewer Application of your choice, and drop them inside the appropriate cubes.

You should at least place on OpenSimulator one box with a 'Bot Controller' LSL script (you can have multiple Bot Controllers if you wish, spread out over several sims or grids, and control them all from a single interface). The Bot Controller item will only launch new agents, it does not do anything else (all AI logic is actually done in Go, and the actual commands given to the agents are directly made from `gobot` to each individual agent). Note that while the Bot Controller will _mostly_ be controlled via the Web interface (or from the `gobot` application itself), you can also _manually_ activate it in-world using chat commands to channel 10, e.g. use `/10 help` to get a list of commands.

Touching the Bot Controller on any lateral face (_not_ the top one!) will reset it, and it will attempt to contact the `gobot` application to do a new registration. The inventory contents will also be uploaded (this is mostly because in the future animations and the like will be controlled from the Bot Controller).

Agents will try to use the Energy, Happiness, and Money cubes; you can have as many of those as you wish, and they are registered with the 'Register Position' script. It's not obvious, but the Description field will tell which class of agent should use that particular cube. I think that the default is 'peasant' but you will have to figure it out on your own (if you leave the description blank, the cube will _not_ work!). The three types of cubes will have hovertext to show its current values, and you can set them up as you wish (there is just one script for all three types anyway); currently, you can touch & drag on the _top_ face to set the three values, although for _precise_ values you might just write the numbers on the description. As with the Bot Controller, you can touch on the _lateral_ faces to reset the cube (it will attempt to register again, send the current settings for Energy/Happiness/Money, get an URL for communication, and so forth). Cubes will regularly update their status with the `gobot` application, just to make sure that the engine has their current positioning data correctly.

Note that you ought to make your cubes and Master Bot Controllers phantom, unless you wish deliberately to have the agents avoid bumping into them. I also expect to create an option to turn them all invisible and/or phantom from a command on the backoffice, for the purposes of shooting videos without the annoying 'control' elements in view.

The tough bit will be configuring Agents, or, as OpenSimulator calls them, NPCs.

Because everything is controlled via `gobot` (as opposed to a specific, C#/libopenmetaverse-based application), NPCs need to communicate (and receive commands) from the exterior world. The script `Register Agent` will basically act as a 'remote controller' for the NPC — it will register with `gobot` and allow the external application to send commands to it. The other two scripts, with the (legacy) names `Sensorama.lsl` and `llCastRay detector script 3.1.lsl` are the NPC's eyes & ears — they will attempt to gather as much data from the environment around the NPC as possible with the limited functionality offered by LSL.

`Sensorama.lsl` is the oldest-fashioned way of collecting in-world objects: it registers a LSL sensor around the NPC, and every item it finds in the sensor's range will be placed into an array. Unfortunately, LSL just sends you back the first 16 items (or agents) it finds. The good news is that at least you get their position, rough size (a bounding box around the item), and even its velocity, if it's moving. The bad news, of course, is that you can miss a lot — and that's why this is a _repeating_ sensor, sending _everything_ that it 'sees' back to the engine on `gobot`. The problem is getting rid of stale objects, i.e. those who have really disappeared from the region (and not merely moved away!), so there is a garbage collector built-in on `gobot` which will delete old objects after a while. There is no problem if it accidentally deletes an 'active' object from the database, which, for some reason, has stopped contacting the application (maybe because the region was slow, or a particular script had low priority when running, or a glitch in the region code, whatever...); later, as that object gets scanned again, it gets back into the database.

Note that scanning objects is a _collaborative_ effort: _all_ agents write to the _same_ database about what items they have 'found'. So, the more agents are roaming the region, the more likely the data will be more and more accurate.

You can (or rather, you should) also place the `Sensorama.lsl` script inside all your cubes, especially those near several complex (static) objects. This will prevent those objects to disappear from the database (when the garbage collector runs) since the cubes will continuously refresh them.

`llCastRay detector script 3.1.lsl` also tries to detect the environment around the NPC. The concept, however, is slightly different: a 'virtual ray' is emitted from the NPC (the direction can be specified in LSL), and everything that this beam crosses will be retrieved into a list of objects — which, in turn, can be polled for additional information (e.g. size, velocity, and so forth). The `llCastRay()` function is directly tied to lower-level tools in the physics engine which provide this functionality directly, as opposed to the sensors, which require geofencing code to extract relevant data from the database (which is more resource-consuming, takes longer, and so forth); thus, `llCastRay()` would _theoretically_ be much more efficient, but there is a catch: it can be 'abused' to the point that the physics engine may be overloaded with way too many calls. Therefore, both OpenSimulator and Second Life limit the amount of rays that can be cast; and once that amount is reached, NPCs would become 'blind', unable to perceive anything in their immediate surroundings. That's why both kind of 'detection' work simultaneously (and cooperatively!) to fill up the database with as much data as it can be extracted from the environment; this 'common knowledge database' is actually what the engine uses to have an idea of how the environment looks like. Note that only the Energy/Happiness/Money are _active_ elements in the environment, telling the engine their _precise_ location; everything else must somehow be acquired with 'artificial vision' through those two scripts.

Note that when the `llCastRay()` detector is operational, there will be a 'laser beam' emitted from the NPC to the item it is currently analysing and considers to be 'in front of them'. Sometimes the NPC will turn the beam off: this means that there are no objects _directly_ in front of it. Unlike sensors, which are cones (or even spheres!), the `llCastRay()` is a tight beam (theoretically of zero thickness) and may therefore produce unexpected results as it goes through certain objects (by narrowly missing their edges!). Nevertheless, _when_ it crosses an object, it is much more accurate in knowing that such an object is _really_ in the path of the agent.

In order for all the above to work, these three scripts (registration/commands, sensor, `llCastRay()`) must be on an object attached to the NPC. So to prepare a 'template' NPC, you should do the following:

  - Start with your own avatar, stripping it from anything unnecessary.
  - Make sure that all scripts have been set to all permissions.
  - Create a transparent object attached to your chest or hip which covers the whole avatar. Cubic designs work better than spheres, but you are welcome to experiment a bit. My own personal choices is a cubic thingy which covers the avatar like a 'bounding box'. It does not need to be a 'perfect' shape — at least, not until a few bugs are fixed to allow NPCs a _third_ way of detecting objects, namely, _bumping into things_; this is supposed to be a 'feature' which works terribly both on SL and on OpenSimulator, so I've mostly abandoned hope of seeing it fixed (and yes, I tried to fix it myself, to no avail), but after seven years of waiting, I simply cannot rely on it any longer — but it works best if most of the avatar is really covered, from head to toe.
  - Once you place the 3 scripts inside, of course they will start registering with `gobot` and send sensor/`llCastRay()` data to the engine; don't worry, you can always delete that later through the Web backoffice (or keep that data, since it will be valid for the environment you're on!).
  - Now comes the tricky bit: your own avatar needs to be 'cloned' to act as a 'template' for the NPCs. This is accomplished by going near to the Bot Controller and typing `/10 clone name-of-notecard`, which, after a few seconds, will drop a notecard with that name _inside the Bot Controller_.
  - Once that's done, you can start creating as many bots as you wish (until the region crashes!) using the Web command to create a new NPC. The first parameter (string) is the new name of the 'bot, and the second parameter (also a string) is the  name of the notecard to be used to create the avatar.
  - If you did all the above steps correctly, you should now be able to create new NPCs which have a copy of the scripts inside a transparent object attached to them, and all will try to register with `gobot` and starting happily to send information about the environment to the engine.

There is some resilience built in the system. If the `gobot` application crashes, the in-world objects (cubes) and the agents will not be able to communicate and update their position; most likely they will stop after a few seconds (not having independent instructions to move). Once the application resumes, some of those objects might have timed-out and given up on the application; they will only attempt a reconnect after an hour or so.

Similarly, a garbage collector will try to refresh the whole database so that it remains more-or-less in sync with the actual items in-world. This means that every 4 hours (currently not configurable!) the garbage collector will attempt to contact with a `ping` all _active_ in-world objects; those that respond to the `ping` are deemed active, the rest is deleted. If the garbage collector was being too aggressive, and just timed out on a `ping` to an object which was active but replied too slowly, that is not a major issue, since after an hour or so those objects will register back again. _Passive_ objects (those which are obstacles and which have been collectively 'detected' by bots and active cubes) are a different story. Some will be relatively static and remain around for a long time; others (like user avatars passing through) might be dynamic and short-lived (i.e. a user might have just logged in and logged out after a few minutes). Because we're not cheating, the agents do not 'know' if an object that 'disappeared' is still there and undetected, or if it really went away. The garbage collector will simply get rid of all objects which have been detected after a certain time has elapsed. The theory is that static objects will quickly be put back into the database again, since the sensors are repeating in seconds; dynamic objects will be caught moving, again, thanks to the many sensors; but those objects or avatars which really went away will never be detected any longer and will not clutter the database with useless junk.

More to come...

## Hidden Easter Eggs

Changing the configuration notecard and saving it will get the application to reload it, and changes will take immediate effect; there is no need to kill the application for that. This allows to change the display severity of the logging on the go, as well as starting/stopping the engine and increasing its verbosity.

Sending a _SIGHUP_ will also try to stop the engine. Note that this might not happen immediately: the engine might run one whole iteration of the whole genetic algorithm before stopping; and messages queued to be sent in-world will probably continue to be sent until the buffer flushes out.

Sending a _SIGCONT_ will try to start the engine (again, if it had stopped before). This might also not happen immediately, although it should be much faster than _stopping_ the engine. Note that in an Unix shell, when you do a Ctrl-Z to stop the application and return to the shell, this will _also_ stop the engine; when you push the application into either the foreground (`fg %1`) or background (`bg %1`), the engine will resume — this is because allegedly Unix will send a _SIGHUP_ with a Ctrl-Z, and a _SIGCONT_ when the application is placed back into running mode (either in the foreground or background). So, at least regarding the engine, `gohup` is compliant with the expected behaviour of an Unix application. Remember, just because the engine might not be running, that doesn't mean that the rest of the application isn't working! (like the whole backoffice interface or the garbage collector — they will run independently from the engine).

_SIGUSR1_ will send a randomly generated female name to appear on the Engine page; _SIGUSR2_ will place a random country. Why? Well, this was the only way to get *something* to appear there while testing the code (and without creating a new backoffice just for testing purposes). Now if I only came up with pipes and more esoteric stuff... 😉

The current version also allows the simulation to run just one step, and then stop (similar to how the old PHP code worked, just one step at the time). This allows for easier debugging (the logs are huge!).

## License

This is currently copyrighted until I finish my PhD, then you can copy it at will. If you wish to use my code **now**, then please ask me before! Thank you!