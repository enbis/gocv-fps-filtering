package main

import (
	"errors"
	"fmt"
	"net/http"

	_ "net/http/pprof"

	log "github.com/sirupsen/logrus"

	"github.com/enbis/gocv-fps-filter/utils"
	"github.com/hybridgroup/mjpeg"
	"github.com/spf13/viper"
	"gocv.io/x/gocv"
)

var (
	deviceID int
	err      error
	webcam   *gocv.VideoCapture
	stream   *mjpeg.Stream
)

func main() {
	err := loadConfig()
	if err != nil {
		log.Error(err.Error())
	}
	startProcess()
}

func loadConfig() error {
	viper.SetDefault("host", "0.0.0.0:8080")
	viper.SetDefault("required_fps", 1)
	viper.SetDefault("video_codec", "MJPG")
	viper.SetDefault("device_id", "1")

	viper.AddConfigPath("./")
	viper.AddConfigPath("./config")
	viper.SetConfigName("config")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		log.Printf("Using config file: %s", viper.ConfigFileUsed())
	}

	return nil
}

func startProcess() {

	deviceID := viper.GetString("device_id")
	host := viper.GetString("host")

	// open webcam
	webcam, err = gocv.OpenVideoCapture(deviceID)

	if err != nil {
		fmt.Printf("Error opening capture device: %v\n", deviceID)
		return
	}
	defer webcam.Close()

	// create the mjpeg stream
	stream = mjpeg.NewStream()
	stream.FrameInterval = 0

	// start capturing
	go mjpegCapture()

	fmt.Println("Capturing. Point your browser to " + host)

	// start http server
	http.Handle("/", stream)
	log.Fatal(http.ListenAndServe(host, nil))
}

func mjpegCapture() {
	img := gocv.NewMat()
	defer img.Close()

	webcam.Set(gocv.VideoCaptureFOURCC, toFOURCC(viper.GetString("video_codec")))

	requiredFps := viper.GetFloat64("required_fps")
	fps := webcam.Get(gocv.VideoCaptureFPS)

	counter := utils.NewCounter(0)
	prop, err := processFps(fps, requiredFps)
	if err != nil {
		log.Error(err.Error())
	}

	iFps := 0
	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("Device closed: %v\n", deviceID)
			return
		}
		if img.Empty() {
			continue
		}

		fmt.Printf("counter %d e iFPS %d \n", counter.GetCount(), iFps)

		if counter.GetCount() == int(fps) {
			counter.SetCounter(0)
			iFps = 0
			continue
		} else if counter.GetCount() == iFps {
			buf, _ := gocv.IMEncode(".jpg", img)
			stream.UpdateJPEG(buf)
			iFps += prop
		} else {
			counter.Increment()
			continue
		}
		counter.Increment()

	}
}

func processFps(running, required float64) (int, error) {
	prop := int(running / required)
	if prop < 1 {
		return 0, errors.New("fps over than running")
	}
	return prop, nil
}

func toFOURCC(codec string) float64 {

	c1 := []rune(string(codec[0]))[0]
	c2 := []rune(string(codec[1]))[0]
	c3 := []rune(string(codec[2]))[0]
	c4 := []rune(string(codec[3]))[0]

	return float64((c1 & 255) + ((c2 & 255) << 8) + ((c3 & 255) << 16) + ((c4 & 255) << 24))
}
