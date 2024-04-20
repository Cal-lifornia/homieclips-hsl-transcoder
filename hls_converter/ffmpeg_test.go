package hls_converter

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/suite"
)

const TestMasterM3u8 = `#EXTM3U
#EXT-X-VERSION:6
#EXT-X-STREAM-INF:BANDWIDTH=22105600,RESOLUTION=3840x2160,CODECS="avc1.640034,mp4a.40.2"
test_object_0.m3u8

#EXT-X-STREAM-INF:BANDWIDTH=11105600,RESOLUTION=2560x1440,CODECS="avc1.640033,mp4a.40.2"
test_object_1.m3u8

#EXT-X-STREAM-INF:BANDWIDTH=5552800,RESOLUTION=1920x1080,CODECS="avc1.640032,mp4a.40.2"
test_object_2.m3u8`

const TestResultM3u8 = `#EXTM3U
#EXT-X-VERSION:6
#EXT-X-STREAM-INF:BANDWIDTH=22105600,RESOLUTION=3840x2160,CODECS="avc1.640034,mp4a.40.2"
test_object_0.m3u8

#EXT-X-STREAM-INF:BANDWIDTH=11105600,RESOLUTION=2560x1440,CODECS="avc1.640033,mp4a.40.2"
test_object_1.m3u8

#EXT-X-STREAM-INF:BANDWIDTH=5552800,RESOLUTION=1920x1080,CODECS="avc1.640032,mp4a.40.2"
test_object_2.m3u8`

type HlsTestSuite struct {
	suite.Suite
	objectName string
}

func (hts *HlsTestSuite) SetupSuite() {
	hts.Run("TestFFmpegExists", func() {
		cmd := exec.Command("ffmpeg", "-version")
		_, err := cmd.Output()
		hts.Require().NoError(err, "ffmpeg should be accessible on command line")
	})

	hts.objectName = "test_object"

	err := os.WriteFile(hts.objectName+"_master.m3u8", []byte(TestMasterM3u8), 0777)
	hts.Require().NoError(err, "Error creating test master file when there shouldn't be")

	hts.Require().FileExistsf(hts.objectName+"_master.m3u8", "master m3u8 should exist")
}

func (hts *HlsTestSuite) TearDownSuite() {
	err := os.Remove(hts.objectName + ".m3u8")
	hts.Require().NoError(err, "Error removing results when there shouldn't be")
}

func (hts *HlsTestSuite) TestEditMasterHls() {
	expectedContents := TestResultM3u8

	//err := EditMasterHls(hts.objectName)
	//hts.Assert().NoError(err, "Error running EditMasterHls when there shouldn't be")

	if hts.Assert().FileExistsf(hts.objectName+".m3u8", hts.objectName+".m3u8 should exist") {
		hts.T().Logf("%s.m3u8 exists", hts.objectName)
		hts.Run("TestResultContentsAreEqual", func() {
			resultContents, err := os.ReadFile(hts.objectName + ".m3u8")
			hts.Require().NoError(err, "should be no error opening %s.m3u8", hts.objectName)
			hts.Assertions.Equal(expectedContents, string(resultContents), "file contents should be equal")
		})
	}
}

func TestHlsTestSuite(t *testing.T) {
	suite.Run(t, new(HlsTestSuite))
}
