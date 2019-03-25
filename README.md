# corrupter
Simple image glitcher suitable for producing nice looking i3lock backgrounds

## Getting Started

```shell
$ git clone https://github.com/r00tman/corrupter
$ cd corrupter && go build
$ ./corrupter -h
$ ./corrupter shots/test2.png out.png && xdg-open out.png
```

At the moment, you can only pass and output png images. But that's enough to work well with scrot and i3lock.

Also, now there's [corrupter-git](https://aur.archlinux.org/packages/corrupter-git/) AUR package maintained by [alrayyes](https://github.com/alrayyes)!

### Less distorted image

Default config is pretty heavy-handed. To get less disrupted images you may want to reduce blur and distortion:
```shell
$ ./corrupter -mag 1 -boffset 2 shots/test2.png out.png && xdg-open out.png
```

### Using with i3lock+scrot/swaylock+grim
Example lock script:
```bash
#!/usr/bin/env bash
tmpbg="/tmp/screen.png"
scrot "$tmpbg"; corrupter "$tmpbg" "$tmpbg"
i3lock -i "$tmpbg"; rm "$tmpbg"
```

## Examples

All images are obtained using the default parameters.
![demo1](https://raw.githubusercontent.com/r00tman/corrupter/master/shots/test2_out.png)
![demo2](https://raw.githubusercontent.com/r00tman/corrupter/master/shots/screen2.png)
![demo3](https://raw.githubusercontent.com/r00tman/corrupter/master/shots/screen5.png)
