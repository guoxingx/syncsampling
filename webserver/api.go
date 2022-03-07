package webserver

import (
	"fmt"
	"net/http"
	"sort"
	"sync/atomic"

	"syncsampling/utils"
	hw "syncsampling/utils/httpwrap"
)

var (
	Images sort.StringSlice
	Index  int32
	Total  int32
)

func handleAction(w http.ResponseWriter, r *http.Request) (int, interface{}) {
	if r.Method != "GET" {
		return 405, nil
	}

	typ := Action(hw.ParseRequestURLArgsUint64(r, "type"))

	if typ == ActionStart {
		// start gallery, reset Index
		logger.Info("receive start signal from web")
		atomic.StoreInt32(&Index, 0)
		return 200, hw.NewRestSuccess(nil)
	} else if typ == ActionReady {
		if Index < Total {
			logger.Info("receive image ready signal from web")
			actionCh <- 1
			return 200, hw.NewRestSuccess(nil)
		}
	}
	return 200, hw.NewRestError(codeNotPlemented, "")
}

func handleImage(w http.ResponseWriter, r *http.Request) (int, interface{}) {
	if r.Method != "GET" {
		return 405, nil
	}

	// index := int(hw.ParseRequestURLArgsUint64(r, "index"))

	if loadImages() != nil {
		return 200, hw.NewRestError(codeServerError, "failed to load images")
	}

	if len(Images) <= int(Index) {
		return 200, hw.NewRestSuccess("")
	}

	url := fmt.Sprintf("http://%s/images/%s", host, Images[Index])

	return 200, hw.NewRestSuccess(url)
}

func handleImages(w http.ResponseWriter, r *http.Request) (int, interface{}) {
	if r.Method != "GET" {
		return 405, nil
	}

	if loadImages() != nil {
		return 200, hw.NewRestError(codeServerError, "failed to load images")
	}
	return 200, hw.NewRestSuccess(Images)
}

func loadImages() error {
	if Images != nil {
		return nil
	}
	dir := "/Users/wuyi/developer/projects/syncsampling/statics/images"
	res, err := utils.ListFiles(dir)
	if err != nil {
		return err
	}
	Images = sort.StringSlice(res)
	Images = sortFiles(Images)

	logger.Infof("%d images found as %s", len(Images), Images)
	atomic.StoreInt32(&Total, int32(len(Images)))

	return nil
}
