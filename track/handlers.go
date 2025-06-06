package track

import (
	"fmt"
	"net/http"
	"time"

	"github.com/patdeg/common"

	"google.golang.org/appengine/v2/user"
)

func CreateTodayVisitsTableInBigQueryHandler(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
	common.Info(">>>>>>>> CreateTodayVisitsTableInBigQueryHandler")

	isAdmin := user.IsAdmin(c)

	if (r.Header.Get("X-AppEngine-Cron") != "true") && (isAdmin == false) {
		common.Error("Handler called without admin/cron privilege")
		http.Error(w, "Handler called without admin/cron privilege", http.StatusBadRequest)
		return
	}

	today := time.Now().Format("20060102")
	err := createVisitsTableInBigQuery(c, today)
	if err != nil {
		common.Error("Error while creating table %v: %v", today, err)
		http.Error(w, "Error while creating today table: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Table %v created", today)
}

func CreateTomorrowVisitsTableInBigQueryHandler(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
	common.Info(">>>>>>>> CreateTomorrowVisitsTableInBigQueryHandler")

	isAdmin := user.IsAdmin(c)

	if (r.Header.Get("X-AppEngine-Cron") != "true") && (isAdmin == false) {
		common.Error("Handler called without admin/cron privilege")
		http.Error(w, "Handler called without admin/cron privilege", http.StatusBadRequest)
		return
	}

	tomorrow := time.Now().Add(time.Hour*23 + time.Minute*59).Format("20060102")
	err := createVisitsTableInBigQuery(c, tomorrow)
	if err != nil {
		common.Error("Error while creating table %v: %v", tomorrow, err)
		http.Error(w, "Error while creating tomorrow table: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Table %v created", tomorrow)
}

func CreateTodayEventsTableInBigQueryHandler(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
	common.Info(">>>>>>>> CreateTomorrowEventsTableInBigQueryHandler")

	isAdmin := user.IsAdmin(c)

	if (r.Header.Get("X-AppEngine-Cron") != "true") && (isAdmin == false) {
		common.Error("Handler called without admin/cron privilege")
		http.Error(w, "Handler called without admin/cron privilege", http.StatusBadRequest)
		return
	}

	today := time.Now().Format("20060102")
	err := createEventsTableInBigQuery(c, today)
	if err != nil {
		common.Error("Error while creating table %v: %v", today, err)
		http.Error(w, "Error while creating today table: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Table %v created", today)
}

func CreateTomorrowEventsTableInBigQueryHandler(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
	common.Info(">>>>>>>> CreateTomorrowEventsTableInBigQueryHandler")

	isAdmin := user.IsAdmin(c)

	if (r.Header.Get("X-AppEngine-Cron") != "true") && (isAdmin == false) {
		common.Error("Handler called without admin/cron privilege")
		http.Error(w, "Handler called without admin/cron privilege", http.StatusBadRequest)
		return
	}

	tomorrow := time.Now().Add(time.Hour*23 + time.Minute*59).Format("20060102")
	err := createEventsTableInBigQuery(c, tomorrow)
	if err != nil {
		common.Error("Error while creating table %v: %v", tomorrow, err)
		http.Error(w, "Error while creating tomorrow table: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Table %v created", tomorrow)
}

func TrackHandler(w http.ResponseWriter, r *http.Request) {
	common.Info(">>>>>>>> TrackHandler")

	common.Info("c=%v a=%v l=%v v=%v", r.FormValue("c"), r.FormValue("a"), r.FormValue("l"), r.FormValue("v"))
	TrackEvent(w, r, common.GetCookieID(w, r))
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Write([]byte(onePixelPNG))
}

func ClickHandler(w http.ResponseWriter, r *http.Request) {
	common.Info(">>>>>>>> ClickHandler")

	common.Info("c=%v a=%v l=%v v=%v", r.FormValue("c"), r.FormValue("a"), r.FormValue("l"), r.FormValue("v"))
	TrackEvent(w, r, common.GetCookieID(w, r))
	url := r.FormValue("url")
	if url == "" {
		url = "http://www.mygotome.com"
	} else if !common.IsValidHTTPURL(url) {
		common.Error("Invalid redirect URL: %v", url)
		url = "http://www.mygotome.com"
	}
	common.Info("Redirect to %v", url)
	http.Redirect(w, r, url, http.StatusFound)
}
