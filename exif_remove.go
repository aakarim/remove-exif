package main

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/dsoprea/go-exif"
	"github.com/dsoprea/go-jpeg-image-structure"
	"github.com/dsoprea/go-png-image-structure"
)

const (
	JpegMediaType  = "jpeg"
	PngMediaType   = "png"
	OtherMediaType = "other"
)

type MediaContext struct {
	MediaType string
	RootIfd   *exif.Ifd
	RawExif   []byte
	Media     interface{}
}

type IfdEntry struct {
	IfdPath     string      `json:"ifd_path"`
	FqIfdPath   string      `json:"fq_ifd_path"`
	IfdIndex    int         `json:"ifd_index"`
	TagId       uint16      `json:"tag_id"`
	TagName     string      `json:"tag_name"`
	TagTypeId   uint16      `json:"tag_type_id"`
	TagTypeName string      `json:"tag_type_name"`
	UnitCount   uint32      `json:"unit_count"`
	Value       interface{} `json:"value"`
	ValueString string      `json:"value_string"`
}

func main() {
	//filepath := os.Args[1]
	//handleFile("img/28-hex_value.jpg")

	var files []string
	root := "img"
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if path != "img" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	pass := 0
	fail := 0
	for _, file := range files {
		fmt.Println(file)
		if _, err := handleFile(file); err != nil {
			fail += 1
			fmt.Printf(err.Error())
		} else {
			pass += 1
		}
		fmt.Println()
	}

	fmt.Printf("Results (%v%%): %v pass, %v fail \n", pass / len(files)*100, pass, fail)

}

func handleFile(filepath string) ([]byte, error) {
	if data, err := ioutil.ReadFile(filepath); err != nil {
		fmt.Printf(err.Error())
		return nil, err
	} else {
		_, err = jpeg.Decode(bytes.NewReader(data))
		if err != nil {
			fmt.Printf("ERROR: original image is corrupt" + err.Error() + "\n")
			return nil, err
		}
		filtered, err := extractEXIF(data)
		if err != nil {
			if !strings.EqualFold(err.Error(), "no exif data") {
				fmt.Printf("* " + err.Error() + "\n")
				return nil, errors.New(err.Error())
			}
		}
		return filtered, nil
	}
}

func extractEXIF(data []byte) ([]byte, error) {
	jmp := jpegstructure.NewJpegMediaParser()
	pmp := pngstructure.NewPngMediaParser()
	mc := &MediaContext{
		MediaType: OtherMediaType,
		RootIfd:   nil,
		RawExif:   nil,
		Media:     nil,
	}
	filtered := []byte{}

	if jmp.LooksLikeFormat(data) {
		mc.MediaType = JpegMediaType
		sl, err := jmp.ParseBytes(data)
		if err != nil {
			return nil, err
		}
		mc.Media = sl

		if rootIfd, rawExif, err := sl.Exif(); err != nil {
			return nil, err
		} else {
			mc.RootIfd = rootIfd
			mc.RawExif = rawExif
		}

		if _, sExif, err := sl.FindExif(); err != nil {
			return nil, err
		} else {
			//fmt.Printf("****(exif) %x %s %x\n", sExif.Offset, sExif.MarkerName, len(sExif.Data))

			bytesCount := 0
			startExifBytes := 4
			endExifBytes := 4
			for _, s := range sl.Segments() {

				if s.MarkerName == sExif.MarkerName {
					if startExifBytes == 4 {
						startExifBytes = bytesCount
						endExifBytes = startExifBytes + len(s.Data)
					} else {
						endExifBytes += len(s.Data)
					}
				}
				bytesCount += len(s.Data)

				//fmt.Printf("%x %s %v (%x)\n", s.Offset, s.MarkerName, len(s.Data), s.Offset+len(s.Data))

			}

			filtered = data
			//filtered = data[:startExifBytes]
			//filtered = append(filtered, data[endExifBytes:]...)

			//fmt.Printf("********(size) %v %v  (%v)\n", len(data), len(filtered), len(data)-len(filtered))

			_, _, err = image.Decode(bytes.NewReader(filtered))
			if err != nil {
				return nil, errors.New("EXIF removal corrupted " + err.Error() + "\n")
			}

		}

	} else if pmp.LooksLikeFormat(data) {
		mc.MediaType = PngMediaType
	}

	return filtered, nil
}