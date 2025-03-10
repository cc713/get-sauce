package hentai2w

import (
	"regexp"
	"strings"

	"github.com/gan-of-culture/get-sauce/request"
	"github.com/gan-of-culture/get-sauce/static"
	"github.com/gan-of-culture/get-sauce/utils"
)

const site = "https://hentai2w.com/"

var reSourceURL = regexp.MustCompile(`<source.*src="([^"]+)"`)

type extractor struct{}

// New returns a hentai2w extractor.
func New() static.Extractor {
	return &extractor{}
}

func (e *extractor) Extract(URL string) ([]*static.Data, error) {
	URLs := parseURL(URL)
	if len(URLs) == 0 {
		return nil, static.ErrURLParseFailed
	}

	data := []*static.Data{}
	for _, u := range URLs {
		d, err := extractData(u)
		if err != nil {
			return nil, utils.Wrap(err, u)
		}
		data = append(data, d)
	}

	return data, nil
}

func parseURL(URL string) []string {
	if strings.HasPrefix(URL, site+"video/") {
		return []string{URL}
	}

	htmlString, err := request.Get(URL)
	if err != nil {
		return []string{}
	}

	re := regexp.MustCompile(`[^"]*/video/[^"]*`)
	return re.FindAllString(htmlString, -1)
}

func extractData(URL string) (*static.Data, error) {
	htmlString, err := request.Get(URL)
	if err != nil {
		return nil, err
	}

	videoURL := utils.GetLastItemString(reSourceURL.FindStringSubmatch(htmlString))
	if videoURL == "" || strings.HasPrefix(videoURL, "<") {
		return nil, static.ErrDataSourceParseFailed
	}
	ext := utils.GetLastItemString(strings.Split(videoURL, "."))

	size, _ := request.Size(videoURL, URL)

	return &static.Data{
		Site:  site,
		Title: utils.GetMeta(&htmlString, "og:title"),
		Type:  "video",
		Streams: map[string]*static.Stream{
			"0": {
				Type: static.DataTypeVideo,
				URLs: []*static.URL{
					{
						URL: videoURL,
						Ext: ext,
					},
				},
				Size: size,
			},
		},
		URL: URL,
	}, nil
}
