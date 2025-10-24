package provider

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/willie68/go_mapproxy/internal/logging"
	"github.com/willie68/go_mapproxy/internal/model"
)

type tmsProvider struct {
	log    *logging.Logger
	config Config
	isTMS  bool
	cl     *http.Client
}

func (s *tmsProvider) Tile(tile model.Tile) (io.ReadCloser, error) {
	tmsURL := s.buildTMSUrl(tile)
	s.log.Debugf("Requesting TMS tile from %s", tmsURL)
	if s.cl == nil {
		s.cl = &http.Client{}
	}
	req, err := http.NewRequest("GET", tmsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	setDefaultHeaders(req)
	for key, value := range s.config.Headers {
		req.Header.Set(key, value)
	}
	resp, err := s.cl.Do(req)
	if err != nil || resp.StatusCode != 200 {
		if resp != nil {
			if resp.Body != nil {
				defer resp.Body.Close()
				bodyBytes, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, errors.New("read error")
				}
				bodyString := string(bodyBytes)
				s.log.Errorf("body: %s", bodyString)
			}
			s.log.Errorf("error on wms request, status: %s: %v", resp.Status, err)
		}
		return nil, fmt.Errorf("Tile error: %v", err)
	}
	return resp.Body, nil
}

func (s *tmsProvider) buildTMSUrl(tile model.Tile) string {
	if s.isTMS {
		// TMS Y coordinate conversion
		ymax := 1 << tile.Z
		tile.Y = ymax - tile.Y - 1
	}
	return fmt.Sprintf("%s/%d/%d/%d.png", s.config.URL, tile.Z, tile.X, tile.Y)
}

// tileToBBox converts TMS tile coordinates to a bounding box in EPSG:3857
//var ymax = 1 << zoom
//var y = ymax - y - 1
