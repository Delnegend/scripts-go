# Scripts-go

## Prerequisites
- BatchAVIF:
  - ffmpeg
  - ffprobe
  - aomenc (or modify the script to use your encoder of choice)
- BatchCompress:
  - 7z
- BatchJXL:
  - cjxl
- BatchResize, UpscaleAni:
  - ffmpeg
  - ffprobe
  - [realesrgan-ncnn-vulkan](https://github.com/xinntao/Real-ESRGAN)
## Parameters

- BatchResize:
  - `dest_size`: the first letter is to select the edge (w, h or a (auto) to select the longest side) to set the max size of the image, followed by the target size.
    - Example: `w100` will upres the image to 100px wide, `h100` will upres the image to 100px high, `a100` will upres the image to 100px wide or high, whichever is the longest side.
    - For ratio instead of pixel size, use `r` followed by the ratio. Example: `r1.5` will upres the image to 150% of its original size.