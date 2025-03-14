package ohentai

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/gan-of-culture/get-sauce/request"
	"github.com/gan-of-culture/get-sauce/static"
	"github.com/gan-of-culture/get-sauce/utils"
)

type source struct {
	File string
}

const site = "https://ohentai.org/"

var reSources = regexp.MustCompile(`\[{.+}\]`)

type extractor struct{}

// New returns a nhentai extractor.
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

// parseURL data
func parseURL(URL string) []string {
	re := regexp.MustCompile(`detail.php\?vid=[^\s']+`)
	URLPart := re.FindString(URL)
	if URLPart != "" {
		return []string{URLPart}
	}

	htmlString, err := request.Get(URL)
	if err != nil {
		return nil
	}

	matchedURLs := re.FindAllString(htmlString, -1)
	return utils.RemoveAdjDuplicates(matchedURLs)[1:]
}

// extractData of URL
func extractData(URL string) (*static.Data, error) {
	URL = site + URL
	htmlString, err := request.Get(URL)
	if err != nil {
		return nil, err
	}

	title := strings.TrimSpace(utils.GetH1(&htmlString, -1))

	matchedSources := reSources.FindString(htmlString)

	sourceInfo := []source{}
	err = json.Unmarshal([]byte(matchedSources), &sourceInfo)
	if err != nil {
		return nil, err
	}

	streams := map[string]*static.Stream{}
	for i, s := range sourceInfo {
		u, err := url.Parse(s.File)
		if err != nil {
			return nil, err
		}

		size, _ := request.Size(s.File, URL)

		streams[fmt.Sprint(i)] = &static.Stream{
			Type: static.DataTypeVideo,
			URLs: []*static.URL{
				{
					URL: s.File,
					Ext: utils.GetFileExt(u.Path),
				},
			},
			Size: size,
		}
	}

	return &static.Data{
		Site:    site,
		Title:   title,
		Type:    static.DataTypeVideo,
		Streams: streams,
		URL:     URL,
	}, nil
}
