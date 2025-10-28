package provider

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/willie68/go_mapproxy/internal/logging"
	"github.com/willie68/go_mapproxy/internal/mercantile"
	"github.com/willie68/go_mapproxy/internal/model"
)

type wmsProvider struct {
	name   string
	log    *logging.Logger
	config Config
	cl     *http.Client
}

func (s *wmsProvider) Tile(tile model.Tile) (io.ReadCloser, error) {
	wmsURL := s.buildWMSUrl(s.tileToBBox(tile))
	s.log.Debugf("Requesting WMS tile from %s", wmsURL)

	if s.cl == nil {
		s.cl = &http.Client{}
	}
	req, err := http.NewRequest("GET", wmsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	setDefaultHeaders(req)
	for key, value := range s.config.Headers {
		req.Header.Set(key, value)
	}

	resp, err := s.cl.Do(req)
	if err != nil || resp.StatusCode != 200 {
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
		return nil, errors.New("Tile error")
	}
	return resp.Body, nil
}

func (s *wmsProvider) buildWMSUrl(bb mercantile.Bbox) string {

	base, err := url.Parse(s.config.URL)
	if err != nil {
		panic(err)
	}

	params := url.Values{}
	params.Add("request", "GetMap")
	params.Add("layers", s.config.Layers)
	params.Add("format", s.config.Format)
	params.Add("bbox", fmt.Sprintf("%.9f,%.9f,%.9f,%.9f", bb.Left, bb.Bottom, bb.Right, bb.Top))
	params.Add("width", "256")
	params.Add("height", "256")
	params.Add("srs", "EPSG:3857")
	if s.config.Version != "" {
		params.Add("version", s.config.Version)
	} else {
		params.Add("version", "1.3.0")
	}
	params.Add("styles", s.config.Styles)

	base.RawQuery = params.Encode()
	wmsURL := base.String()

	s.log.Debugf("wms url: %s", wmsURL)
	return wmsURL
}

// Hilfsfunktion für XYZ->BBOX-Konvertierung
func (s *wmsProvider) tileToBBox(tile model.Tile) mercantile.Bbox {
	t := mercantile.TileID{
		X: tile.X,
		Y: tile.Y,
		Z: int(tile.Z),
	}
	return mercantile.XyBounds(t)
}
