# Corrupter
### Simple image glitcher suitable for producing nice looking i3lock backgrounds

## Getting Started

```shell
$ git clone https://github.com/r00tman/corrupter
$ cd corrupter && go build
$ ./corrupter -h
$ ./corrupter shots/test2.png out.png && xdg-open out.png
```

If you're using arch-based there's 2 AUR packages! \
[corrupter-git](https://aur.archlinux.org/packages/corrupter-git/) maintained by [alrayyes](https://github.com/alrayyes), for an automatic build and \
[corrupter-bin](https://aur.archlinux.org/packages/corrupter-git/) maintained by [marcospb19](https://github.com/marcospb19) for the pre-built binary

At the moment, you can only pass and output `.png` images. But that's enough to work well with `scrot` and `i3lock`.

### Using with i3lock+scrot / swaylock+grim
As corrupter only glitches the image for a cool background, you'll have to set up a lock script

Example screenshot lock script:
```bash
#!/usr/bin/env bash
tmpbg="/tmp/screen.png"
scrot "$tmpbg"; corrupter "$tmpbg" "$tmpbg"
i3lock -i "$tmpbg"; rm "$tmpbg"
```

The script above takes a screenshot with `scrot`, distorts it with `corrupter`, and then locks the screen using `i3lock`

If you're using `i3`, you can create the script at `~/.corrupter`, and then use a lock `bindsym`
```
bindsym $mod+l exec --no-startup-id ./.corrupter

```

### Using pre-corrupted images
Alternatively, you can pre-corrupt an image and always use it (which is faster)
```shell
$ ./corrupter shots/test2.png ~/.wallpaper.png
```

and then, inside of your `i3/.config`
```
bindsym $mod+l exec --no-startup-id i3lock -i ./.wallpaper.png

```

This method is slightly faster because the image processing is already done


### Less distorted image

Default config is pretty heavy-handed. To get less disrupted images you may want to reduce blur and distortion:
```shell
$ ./corrupter -mag 1 -boffset 2 shots/test2.png out.png && xdg-open out.png
```

## Examples

All images are obtained using the default parameters.
![demo1](https://raw.githubusercontent.com/r00tman/corrupter/master/shots/test2_out.png)
![demo2](https://raw.githubusercontent.com/r00tman/corrupter/master/shots/screen2.png)
![demo3](https://raw.githubusercontent.com/r00tman/corrupter/master/shots/screen5.png)
