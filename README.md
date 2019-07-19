# corrupter
Simple image glitcher suitable for producing nice looking i3lock backgrounds

## Getting Started

```shell
$ git clone https://github.com/r00tman/corrupter
$ cd corrupter && go build
$ ./corrupter -h
$ ./corrupter shots/example.png out.png && xdg-open out.png
```

If you're using an arch-based distro there are 2 AUR packages!
 - [corrupter-git](https://aur.archlinux.org/packages/corrupter-git/) maintained by [alrayyes](https://github.com/alrayyes), for an automated build, and
 - [corrupter-bin](https://aur.archlinux.org/packages/corrupter-bin/) maintained by [marcospb19](https://github.com/marcospb19), for the pre-built binary.

At the moment, you can only pass and output `.png` images. But that's enough to work well with `scrot` and `i3lock`.

### Using with i3lock+scrot / swaylock+grim
As `corrupter` only glitches the image for a cool background, you'd have to set up a lock script.

Example screenshot lock script:
```bash
#!/usr/bin/env bash
tmpbg="/tmp/screen.png"
scrot "$tmpbg"; corrupter "$tmpbg" "$tmpbg"
i3lock -i "$tmpbg"; rm "$tmpbg"
```

The script above takes a screenshot with `scrot`, distorts it with `corrupter`, and then locks the screen using `i3lock`.

If you're using `i3`, you can create the script at `~/.lock`, and then add a lock `bindsym`.
```
bindsym $mod+Control+l exec --no-startup-id bash ./.lock
```

### Using pre-corrupted images
Alternatively, you can pre-corrupt an image and always use it (which is faster):
```shell
$ ./corrupter shots/example.png ~/.wallpaper.png
```

and then, in your `~/.config/i3/config`:
```
bindsym $mod+Control+l exec --no-startup-id i3lock -i ./.wallpaper.png
```

This method is slightly faster since the image processing is already done.


### Less distorted image

Default config is pretty heavy-handed. To get less disrupted images you may want to reduce blur and distortion:
```shell
$ ./corrupter -mag 1 -boffset 2 shots/example.png out.png && xdg-open out.png
```

## Examples

Images using the default parameters:
![demo1](https://raw.githubusercontent.com/r00tman/corrupter/master/shots/example-after.png)
![demo2](https://raw.githubusercontent.com/r00tman/corrupter/master/shots/light-theme-example.png)
![demo3](https://raw.githubusercontent.com/r00tman/corrupter/master/shots/dark-theme-example.png)

With custom parameters: \
Before:
![demo4](https://raw.githubusercontent.com/r00tman/corrupter/master/shots/ps2-example-before.png)

After (custom parameters and ImageMagick dim):
![demo5](https://raw.githubusercontent.com/r00tman/corrupter/master/shots/ps2-example-after.png)
