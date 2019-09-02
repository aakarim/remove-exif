# go-exif-remove
Removes EXIF information from JPG and PNG files

[![Build Status](https://img.shields.io/circleci/project/github/scottleedavis/go-exif-remove/master.svg)](https://circleci.com/gh/scottleedavis/go-exif-remove)

Uses [go-exif](https://github.com/dsoprea/go-exif) to extract EXIF information and overwrites the EXIF region.

```go
import 	"github.com/scottleedavis/go-exif-remove"

noExifBytes, err := exifremove.Remove(imageBytes)
```

See example usage in [exif-remove-tool](exif-remove-tool)

