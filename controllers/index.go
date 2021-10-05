package controllers

import (
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"text/template"

	log "github.com/sirupsen/logrus"

	"github.com/dolfly/owncast/core"
	"github.com/dolfly/owncast/core/data"
	"github.com/dolfly/owncast/models"
	"github.com/dolfly/owncast/router/middleware"
	"github.com/dolfly/owncast/utils"
	"github.com/dolfly/owncast/webroot"
)

// MetadataPage represents a server-rendered web page for bots and web scrapers.
type MetadataPage struct {
	RequestedURL  string
	Image         string
	Thumbnail     string
	TagsString    string
	Summary       string
	Name          string
	Tags          []string
	SocialHandles []models.SocialHandle
}

// IndexHandler handles the default index route.
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	middleware.EnableCors(w)
	isIndexRequest := r.URL.Path == "/" || filepath.Base(r.URL.Path) == "index.html" || filepath.Base(r.URL.Path) == ""

	// For search engine bots and social scrapers return a special
	// server-rendered page.
	if utils.IsUserAgentABot(r.UserAgent()) && isIndexRequest {
		handleScraperMetadataPage(w, r)
		return
	}

	if utils.IsUserAgentAPlayer(r.UserAgent()) && isIndexRequest {
		http.Redirect(w, r, "/hls/stream.m3u8", http.StatusTemporaryRedirect)
		return
	}

	// If the ETags match then return a StatusNotModified
	if responseCode := middleware.ProcessEtags(w, r); responseCode != 0 {
		w.WriteHeader(responseCode)
		return
	}

	// Set a cache control max-age header
	middleware.SetCachingHeaders(w, r)

	// Set our global HTTP headers
	middleware.SetHeaders(w)
	http.FileServer(http.FS(webroot.FS)).ServeHTTP(w, r)
}

// Return a basic HTML page with server-rendered metadata from the config file
// to give to Opengraph clients and web scrapers (bots, web crawlers, etc).
func handleScraperMetadataPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFS(webroot.FS, "*.html"))

	scheme := "http"

	if siteURL := data.GetServerURL(); siteURL != "" {
		if parsed, err := url.Parse(siteURL); err == nil && parsed.Scheme != "" {
			scheme = parsed.Scheme
		}
	}

	fullURL, err := url.Parse(fmt.Sprintf("%s://%s%s", scheme, r.Host, r.URL.Path))
	if err != nil {
		log.Panicln(err)
	}
	imageURL, err := url.Parse(fmt.Sprintf("%s://%s%s", scheme, r.Host, "/logo/external"))
	if err != nil {
		log.Panicln(err)
	}

	status := core.GetStatus()

	// we're offline then just use the logo image
	var thumbnailURL string
	if status.Online {
		thumbnail, err := url.Parse(fmt.Sprintf("%s://%s%s", scheme, r.Host, "/thumbnail.jpg"))
		if err != nil {
			log.Errorln(err)
			thumbnailURL = imageURL.String()
		} else {
			thumbnailURL = thumbnail.String()
		}
	} else {
		thumbnailURL = imageURL.String()
	}

	tagsString := strings.Join(data.GetServerMetadataTags(), ",")
	metadata := MetadataPage{
		Name:          data.GetServerName(),
		RequestedURL:  fullURL.String(),
		Image:         imageURL.String(),
		Summary:       data.GetServerSummary(),
		Thumbnail:     thumbnailURL,
		TagsString:    tagsString,
		Tags:          data.GetServerMetadataTags(),
		SocialHandles: data.GetSocialHandles(),
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.ExecuteTemplate(w, "metadata.html", metadata); err != nil {
		log.Panicln(err)
	}
}
