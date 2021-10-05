package core

import (
	"os"
	"path"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/dolfly/owncast/config"
	"github.com/dolfly/owncast/core/chat"
	"github.com/dolfly/owncast/core/data"
	"github.com/dolfly/owncast/core/rtmp"
	"github.com/dolfly/owncast/core/transcoder"
	"github.com/dolfly/owncast/core/user"
	"github.com/dolfly/owncast/models"
	"github.com/dolfly/owncast/utils"
	"github.com/dolfly/owncast/yp"
)

var (
	_stats       *models.Stats
	_storage     models.StorageProvider
	_transcoder  *transcoder.Transcoder
	_yp          *yp.YP
	_broadcaster *models.Broadcaster
)

var handler transcoder.HLSHandler
var fileWriter = transcoder.FileWriterReceiverService{}

// Start starts up the core processing.
func Start() error {
	resetDirectories()

	data.PopulateDefaults()

	if err := data.VerifySettings(); err != nil {
		log.Error(err)
		return err
	}

	if err := setupStats(); err != nil {
		log.Error("failed to setup the stats")
		return err
	}

	// The HLS handler takes the written HLS playlists and segments
	// and makes storage decisions.  It's rather simple right now
	// but will play more useful when recordings come into play.
	handler = transcoder.HLSHandler{}

	if err := setupStorage(); err != nil {
		log.Errorln("storage error", err)
	}

	user.SetupUsers()

	fileWriter.SetupFileWriterReceiverService(&handler)

	if err := createInitialOfflineState(); err != nil {
		log.Error("failed to create the initial offline state")
		return err
	}

	_yp = yp.NewYP(GetStatus)

	if err := chat.Start(GetStatus); err != nil {
		log.Errorln(err)
	}

	// start the rtmp server
	go rtmp.Start(setStreamAsConnected, setBroadcaster)

	rtmpPort := data.GetRTMPPortNumber()
	log.Infof("RTMP is accepting inbound streams on port %d.", rtmpPort)

	return nil
}

func createInitialOfflineState() error {
	transitionToOfflineVideoStreamContent()
	return nil
}

// transitionToOfflineVideoStreamContent will overwrite the current stream with the
// offline video stream state only.  No live stream HLS segments will continue to be
// referenced.
func transitionToOfflineVideoStreamContent() {
	log.Traceln("Firing transcoder with offline stream state")

	offlineFilename := "offline.ts"
	offlineFilePath := filepath.Join(config.DataDirectory, offlineFilename)
	_transcoder := transcoder.NewTranscoder()
	_transcoder.SetInput(offlineFilePath)
	_transcoder.SetIdentifier("offline")
	_transcoder.Start()

	// Delete the preview Gif
	_ = os.Remove(path.Join(config.DataDirectory, "preview.gif"))
}

func resetDirectories() {
	log.Trace("Resetting file directories to a clean slate.")

	// Wipe hls data directory
	utils.CleanupDirectory(config.HLSStoragePath)

	// Remove the previous thumbnail
	logo := data.GetLogoPath()
	if utils.DoesFileExists(logo) {
		err := utils.Copy(path.Join(config.DataDirectory, logo), filepath.Join(config.DataDirectory, "thumbnail.jpg"))
		if err != nil {
			log.Warnln(err)
		}
	}
}
